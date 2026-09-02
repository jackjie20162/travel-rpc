package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type InventoryService struct {
	travel.UnimplementedInventoryServiceServer
	repo repository.InventoryRepository
}

func NewInventoryService(repo repository.InventoryRepository) *InventoryService {
	return &InventoryService{repo: repo}
}

func (s *InventoryService) Check(ctx context.Context, req *travel.InventoryRequest) (*travel.InventoryResponse, error) {
	if req == nil || req.GetPackageId() <= 0 || req.GetDate() == "" || req.GetQuantity() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "package, date and positive quantity are required")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	item, err := s.repo.Check(ctx, tenantID, req.GetPackageId(), req.GetDate(), req.GetTimeSlot(), int(req.GetQuantity()))
	if err != nil { return nil, status.Error(codes.NotFound, "inventory not found") }
	return &travel.InventoryResponse{Available:item.Remaining >= int(req.GetQuantity()), Remaining:int32(item.Remaining), UnitPrice:item.UnitPrice, Currency:item.Currency}, nil
}

func (s *InventoryService) Reserve(ctx context.Context, req *travel.ReserveInventoryRequest) (*travel.ReserveInventoryResponse, error) {
	if req == nil || req.GetPackageId() <= 0 || req.GetDate() == "" || req.GetQuantity() <= 0 || req.GetReservationKey() == "" {
		return nil, status.Error(codes.InvalidArgument, "package, date, positive quantity and reservation key are required")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	merchantID, err := auth.MerchantID(ctx)
	if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	result, err := s.repo.Reserve(ctx, tenantID, merchantID, req.GetPackageId(), req.GetDate(), req.GetTimeSlot(), int(req.GetQuantity()), req.GetReservationKey())
	if err != nil {
		if _, ok := err.(*repository.ErrInsufficientInventory); ok { return nil, status.Error(codes.ResourceExhausted, "insufficient inventory") }
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &travel.ReserveInventoryResponse{Reserved:true, Remaining:int32(result.Remaining), UnitPrice:result.UnitPrice, Currency:result.Currency}, nil
}
