package repository

import (
    "context"
    "fmt"
    "time"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/ent/order"
    "gitee.com/meinongyihe/travel-rpc/ent/payment"
)

type mysqlPaymentRepository struct { client *ent.Client }
func NewPaymentRepository(client *ent.Client) PaymentRepository { return &mysqlPaymentRepository{client: client} }

func (r *mysqlPaymentRepository) Create(ctx context.Context, input CreatePaymentInput) (*PaymentRecord, error) {
    if input.TenantID <= 0 || input.MerchantID <= 0 || input.OrderNo == "" || input.Provider == "" || input.IdempotencyKey == "" { return nil, &ErrPaymentConflict{} }
    existing, err := r.client.Payment.Query().Where(payment.TenantIDEQ(input.TenantID), payment.MerchantIDEQ(input.MerchantID), payment.IdempotencyKeyEQ(input.IdempotencyKey)).Only(ctx)
    if err == nil { return r.toRecord(existing, input.OrderNo), nil }
    if !ent.IsNotFound(err) { return nil, err }
    o, err := r.client.Order.Query().Where(order.OrderNoEQ(input.OrderNo), order.TenantIDEQ(input.TenantID), order.MerchantIDEQ(input.MerchantID)).Only(ctx)
    if err != nil { return nil, &ErrPaymentNotFound{} }
    if o.Status != "PENDING_PAYMENT" && o.Status != "PAYMENT_PROCESSING" { return nil, &ErrPaymentConflict{} }
    p, err := r.client.Payment.Create().SetTenantID(input.TenantID).SetMerchantID(input.MerchantID).SetOrderID(o.ID).SetPaymentNo(newPaymentNo()).SetProvider(input.Provider).SetAmount(o.TotalAmount).SetCurrency(o.Currency).SetStatus("CREATED").SetIdempotencyKey(input.IdempotencyKey).Save(ctx)
    if err != nil { return nil, err }
    return r.toRecord(p, o.OrderNo), nil
}

func (r *mysqlPaymentRepository) Get(ctx context.Context, tenantID, merchantID int64, paymentNo string) (*PaymentRecord, error) {
    p, err := r.client.Payment.Query().Where(payment.TenantIDEQ(tenantID), payment.MerchantIDEQ(merchantID), payment.PaymentNoEQ(paymentNo)).Only(ctx)
    if err != nil { return nil, &ErrPaymentNotFound{} }
    o, err := r.client.Order.Get(ctx, int(p.OrderID)); if err != nil || o.TenantID != tenantID || o.MerchantID != merchantID { return nil, &ErrPaymentNotFound{} }
    return r.toRecord(p, o.OrderNo), nil
}

func (r *mysqlPaymentRepository) SetProviderID(ctx context.Context, tenantID, merchantID int64, paymentNo, providerPaymentID string) (*PaymentRecord, error) {
    if providerPaymentID == "" { return nil, &ErrPaymentConflict{} }
    p, err := r.client.Payment.Query().Where(payment.TenantIDEQ(tenantID), payment.MerchantIDEQ(merchantID), payment.PaymentNoEQ(paymentNo)).Only(ctx)
    if err != nil { return nil, &ErrPaymentNotFound{} }
    if p.Status == "PAID" && p.ProviderPaymentID != providerPaymentID { return nil, &ErrPaymentConflict{} }
    if p.ProviderPaymentID != "" && p.ProviderPaymentID != providerPaymentID { return nil, &ErrPaymentConflict{} }
    p, err = r.client.Payment.UpdateOneID(p.ID).SetProviderPaymentID(providerPaymentID).SetStatus("PROCESSING").Save(ctx); if err != nil { return nil, err }
    o, err := r.client.Order.Get(ctx, int(p.OrderID)); if err != nil { return nil, err }
    return r.toRecord(p, o.OrderNo), nil
}

func (r *mysqlPaymentRepository) MarkPaid(ctx context.Context, tenantID, merchantID int64, paymentNo, providerPaymentID string) (*PaymentRecord, error) {
    if providerPaymentID == "" { return nil, &ErrPaymentConflict{} }
    tx, err := r.client.Tx(ctx); if err != nil { return nil, err }
    committed := false; defer func() { if !committed { _ = tx.Rollback() } }()
    p, err := tx.Payment.Query().Where(payment.TenantIDEQ(tenantID), payment.MerchantIDEQ(merchantID), payment.PaymentNoEQ(paymentNo)).Only(ctx)
    if err != nil { return nil, &ErrPaymentNotFound{} }
    if p.Status == "PAID" {
        if p.ProviderPaymentID != providerPaymentID { return nil, &ErrPaymentConflict{} }
        o, err := tx.Order.Get(ctx, int(p.OrderID)); if err != nil { return nil, err }
        if err = tx.Commit(); err != nil { return nil, err }; committed = true
        return r.toRecord(p, o.OrderNo), nil
    }
    if p.Status != "CREATED" && p.Status != "PROCESSING" { return nil, &ErrPaymentConflict{} }
    if p.ProviderPaymentID != "" && p.ProviderPaymentID != providerPaymentID { return nil, &ErrPaymentConflict{} }
    o, err := tx.Order.Get(ctx, int(p.OrderID)); if err != nil || o.TenantID != tenantID || o.MerchantID != merchantID { return nil, &ErrPaymentNotFound{} }
    if o.Status != "PENDING_PAYMENT" && o.Status != "PAYMENT_PROCESSING" { return nil, &ErrPaymentConflict{} }
    p, err = tx.Payment.UpdateOneID(p.ID).SetProviderPaymentID(providerPaymentID).SetStatus("PAID").Save(ctx); if err != nil { return nil, err }
    o, err = tx.Order.UpdateOneID(o.ID).SetPaymentStatus("PAID").SetStatus("CONFIRMED").Save(ctx); if err != nil { return nil, err }
    if err = tx.Commit(); err != nil { return nil, err }; committed = true
    return r.toRecord(p, o.OrderNo), nil
}

func (r *mysqlPaymentRepository) toRecord(p *ent.Payment, orderNo string) *PaymentRecord { return &PaymentRecord{ID:int64(p.ID), PaymentNo:p.PaymentNo, OrderNo:orderNo, Provider:p.Provider, ProviderPaymentID:p.ProviderPaymentID, Amount:p.Amount, Currency:p.Currency, Status:p.Status} }
func newPaymentNo() string { return fmt.Sprintf("PAY%d", time.Now().UnixNano()) }
