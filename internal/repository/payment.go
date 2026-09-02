package repository

import "context"

type CreatePaymentInput struct {
    TenantID       int64
    MerchantID     int64
    OrderNo        string
    Provider       string
    IdempotencyKey string
}

type PaymentRepository interface {
    Create(ctx context.Context, input CreatePaymentInput) (*PaymentRecord, error)
    Get(ctx context.Context, tenantID, merchantID int64, paymentNo string) (*PaymentRecord, error)
    MarkPaid(ctx context.Context, tenantID, merchantID int64, paymentNo, providerPaymentID string) (*PaymentRecord, error)
}

type PaymentRecord struct {
    ID                 int64
    PaymentNo          string
    OrderNo            string
    Provider           string
    ProviderPaymentID  string
    Amount             int64
    Currency           string
    Status             string
}

type ErrPaymentNotFound struct{}
func (e *ErrPaymentNotFound) Error() string { return "payment not found" }

type ErrPaymentConflict struct{}
func (e *ErrPaymentConflict) Error() string { return "payment conflict" }
