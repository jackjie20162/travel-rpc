package service

import (
    "context"

    "gitee.com/meinongyihe/travel-rpc/internal/auth"
    "gitee.com/meinongyihe/travel-rpc/internal/repository"
    "gitee.com/meinongyihe/travel-rpc/travel"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type PaymentService struct {
    travel.UnimplementedPaymentServiceServer
    payments repository.PaymentRepository
}

func NewPaymentService(payments repository.PaymentRepository) *PaymentService { return &PaymentService{payments: payments} }

func (s *PaymentService) Create(ctx context.Context, req *travel.CreatePaymentRequest) (*travel.Payment, error) {
    if req == nil || req.GetOrderNo() == "" || req.GetProvider() == "" || req.GetIdempotencyKey() == "" { return nil, status.Error(codes.InvalidArgument, "order number, provider and idempotency key are required") }
    tenantID, err := auth.TenantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    p, err := s.payments.Create(ctx, repository.CreatePaymentInput{TenantID: tenantID, MerchantID: merchantID, OrderNo: req.GetOrderNo(), Provider: req.GetProvider(), IdempotencyKey: req.GetIdempotencyKey()})
    if err != nil { return nil, mapPaymentError(err) }
    return toPayment(p), nil
}

func (s *PaymentService) Get(ctx context.Context, req *travel.PaymentNoRequest) (*travel.Payment, error) {
    if req == nil || req.GetPaymentNo() == "" { return nil, status.Error(codes.InvalidArgument, "payment number is required") }
    tenantID, err := auth.TenantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    p, err := s.payments.Get(ctx, tenantID, merchantID, req.GetPaymentNo())
    if err != nil { return nil, mapPaymentError(err) }
    return toPayment(p), nil
}

func (s *PaymentService) MarkPaid(ctx context.Context, req *travel.MarkPaymentPaidRequest) (*travel.Payment, error) {
    if req == nil || req.GetPaymentNo() == "" || req.GetProviderPaymentId() == "" { return nil, status.Error(codes.InvalidArgument, "payment number and provider payment id are required") }
    tenantID, err := auth.TenantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    p, err := s.payments.MarkPaid(ctx, tenantID, merchantID, req.GetPaymentNo(), req.GetProviderPaymentId())
    if err != nil { return nil, mapPaymentError(err) }
    return toPayment(p), nil
}

func toPayment(p *repository.PaymentRecord) *travel.Payment {
    return &travel.Payment{Id: p.ID, PaymentNo: p.PaymentNo, OrderNo: p.OrderNo, Provider: p.Provider, ProviderPaymentId: p.ProviderPaymentID, Amount: p.Amount, Currency: p.Currency, Status: p.Status}
}

func mapPaymentError(err error) error {
    switch err.(type) {
    case *repository.ErrPaymentNotFound: return status.Error(codes.NotFound, err.Error())
    case *repository.ErrPaymentConflict: return status.Error(codes.Aborted, err.Error())
    default: return status.Error(codes.Internal, err.Error())
    }
}
