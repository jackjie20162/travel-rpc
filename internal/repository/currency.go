package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// BaseCurrencyCode is the settlement/base currency. All product prices and
// order total_amount are stored in this currency (int64 major units).
const BaseCurrencyCode = "AED"

// RateMicroScale is the multiplier used to persist a real exchange rate as an
// int64 (rate_micro = real rate × 1e6), avoiding float precision drift.
const RateMicroScale int64 = 1000000

// CurrencyRepository provides read access to the global currency dictionary
// and exchange rates, plus an idempotent seeder for the built-in currencies.
// Exchange rates are always stored relative to BaseCurrencyCode.
type CurrencyRepository interface {
	// ListCurrencies returns ACTIVE currencies ordered by sort then code.
	ListCurrencies(ctx context.Context) ([]*ent.Currency, error)
	// ListExchangeRates returns the latest ACTIVE rate for each target currency
	// relative to base. An empty base defaults to BaseCurrencyCode.
	ListExchangeRates(ctx context.Context, base string) ([]*ent.ExchangeRate, error)
	// LatestRateMicro resolves the effective rate_micro for base→target.
	// It reports found=false when no ACTIVE rate exists so callers can fall back
	// to the base currency instead of locking an incorrect parity rate.
	LatestRateMicro(ctx context.Context, base, target string) (rateMicro int64, found bool, err error)
	// SeedDefaults idempotently inserts the built-in currencies (AED/USD/CNY)
	// and their default exchange rates when missing.
	SeedDefaults(ctx context.Context) error
}
