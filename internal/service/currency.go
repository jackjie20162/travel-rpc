package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// CurrencyService exposes the global currency dictionary and exchange rates.
// Currency data is not tenant-scoped, so no auth metadata is required.
type CurrencyService struct {
	travel.UnimplementedCurrencyServiceServer
	currencies repository.CurrencyRepository
}

func NewCurrencyService(currencies repository.CurrencyRepository) *CurrencyService {
	return &CurrencyService{currencies: currencies}
}

func (s *CurrencyService) ListCurrencies(ctx context.Context, _ *travel.ListCurrenciesRequest) (*travel.ListCurrenciesResponse, error) {
	items, err := s.currencies.ListCurrencies(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &travel.ListCurrenciesResponse{}
	for _, c := range items {
		resp.Items = append(resp.Items, toCurrency(c))
	}
	return resp, nil
}

func (s *CurrencyService) ListExchangeRates(ctx context.Context, req *travel.ListExchangeRatesRequest) (*travel.ListExchangeRatesResponse, error) {
	base := req.GetBase()
	if base == "" {
		base = repository.BaseCurrencyCode
	}
	items, err := s.currencies.ListExchangeRates(ctx, base)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	resp := &travel.ListExchangeRatesResponse{}
	for _, r := range items {
		resp.Items = append(resp.Items, toExchangeRate(r))
	}
	return resp, nil
}

func toCurrency(c *ent.Currency) *travel.Currency {
	return &travel.Currency{
		Code:           c.Code,
		Symbol:         c.Symbol,
		NameKey:        c.NameKey,
		Decimals:       int32(c.Decimals),
		SymbolPosition: c.SymbolPosition,
		IsBase:         c.IsBase,
	}
}

func toExchangeRate(r *ent.ExchangeRate) *travel.ExchangeRate {
	return &travel.ExchangeRate{
		BaseCurrency:   r.BaseCurrency,
		TargetCurrency: r.TargetCurrency,
		RateMicro:      r.RateMicro,
		EffectiveAt:    r.EffectiveAt,
	}
}
