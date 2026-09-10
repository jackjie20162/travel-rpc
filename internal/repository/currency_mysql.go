package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/currency"
	"gitee.com/meinongyihe/travel-rpc/ent/exchangerate"
)

type mysqlCurrencyRepository struct{ client *ent.Client }

func NewCurrencyRepository(client *ent.Client) CurrencyRepository {
	return &mysqlCurrencyRepository{client: client}
}

func (r *mysqlCurrencyRepository) ListCurrencies(ctx context.Context) ([]*ent.Currency, error) {
	return r.client.Currency.Query().
		Where(currency.StatusEQ("ACTIVE")).
		Order(ent.Asc(currency.FieldSort), ent.Asc(currency.FieldCode)).
		All(ctx)
}

func (r *mysqlCurrencyRepository) ListExchangeRates(ctx context.Context, base string) ([]*ent.ExchangeRate, error) {
	if base == "" {
		base = BaseCurrencyCode
	}
	// Latest effective_at first, then keep only the most recent row per target.
	rows, err := r.client.ExchangeRate.Query().
		Where(exchangerate.BaseCurrencyEQ(base), exchangerate.StatusEQ("ACTIVE")).
		Order(ent.Desc(exchangerate.FieldEffectiveAt), ent.Desc(exchangerate.FieldID)).
		All(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]struct{}, len(rows))
	out := make([]*ent.ExchangeRate, 0, len(rows))
	for _, row := range rows {
		if _, ok := seen[row.TargetCurrency]; ok {
			continue
		}
		seen[row.TargetCurrency] = struct{}{}
		out = append(out, row)
	}
	return out, nil
}

func (r *mysqlCurrencyRepository) LatestRateMicro(ctx context.Context, base, target string) (int64, bool, error) {
	if base == "" {
		base = BaseCurrencyCode
	}
	// Parity when converting to the base currency itself.
	if target == "" || target == base {
		return RateMicroScale, true, nil
	}
	rows, err := r.client.ExchangeRate.Query().
		Where(exchangerate.BaseCurrencyEQ(base), exchangerate.TargetCurrencyEQ(target), exchangerate.StatusEQ("ACTIVE")).
		Order(ent.Desc(exchangerate.FieldEffectiveAt), ent.Desc(exchangerate.FieldID)).
		Limit(1).
		All(ctx)
	if err != nil {
		return 0, false, err
	}
	if len(rows) == 0 || rows[0].RateMicro <= 0 {
		return 0, false, nil
	}
	return rows[0].RateMicro, true, nil
}

// currencySeed describes a built-in currency and its default rate relative to
// the base currency (rate_micro). The base currency itself is seeded at parity.
type currencySeed struct {
	code       string
	symbol     string
	nameKey    string
	decimals   int
	position   string
	isBase     bool
	sort       int
	rateMicro  int64
}

// defaultCurrencySeeds matches the frontend static fallback rates so that the
// UI and backend agree before any operational rate import.
var defaultCurrencySeeds = []currencySeed{
	{code: "AED", symbol: "د.إ", nameKey: "currency.AED", decimals: 2, position: "suffix", isBase: true, sort: 1, rateMicro: RateMicroScale},
	{code: "USD", symbol: "$", nameKey: "currency.USD", decimals: 2, position: "prefix", isBase: false, sort: 2, rateMicro: 272300},  // 1 AED ≈ 0.2723 USD
	{code: "CNY", symbol: "¥", nameKey: "currency.CNY", decimals: 2, position: "prefix", isBase: false, sort: 3, rateMicro: 1960000}, // 1 AED ≈ 1.96 CNY
}

func (r *mysqlCurrencyRepository) SeedDefaults(ctx context.Context) error {
	now := time.Now().Unix()
	for _, seed := range defaultCurrencySeeds {
		exists, err := r.client.Currency.Query().Where(currency.CodeEQ(seed.code)).Exist(ctx)
		if err != nil {
			return err
		}
		if !exists {
			if _, err := r.client.Currency.Create().
				SetCode(seed.code).
				SetSymbol(seed.symbol).
				SetNameKey(seed.nameKey).
				SetDecimals(seed.decimals).
				SetSymbolPosition(seed.position).
				SetIsBase(seed.isBase).
				SetStatus("ACTIVE").
				SetSort(seed.sort).
				Save(ctx); err != nil {
				return err
			}
		}
		// The base currency needs no conversion rate row.
		if seed.code == BaseCurrencyCode {
			continue
		}
		rateExists, err := r.client.ExchangeRate.Query().
			Where(exchangerate.BaseCurrencyEQ(BaseCurrencyCode), exchangerate.TargetCurrencyEQ(seed.code)).
			Exist(ctx)
		if err != nil {
			return err
		}
		if !rateExists {
			if _, err := r.client.ExchangeRate.Create().
				SetBaseCurrency(BaseCurrencyCode).
				SetTargetCurrency(seed.code).
				SetRateMicro(seed.rateMicro).
				SetSource("MANUAL").
				SetEffectiveAt(now).
				SetStatus("ACTIVE").
				Save(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}
