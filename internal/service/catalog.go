package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService struct { travel.UnimplementedCatalogServiceServer; repo repository.ProductRepository }
func NewCatalogService(repo repository.ProductRepository) *CatalogService { return &CatalogService{repo:repo} }

func (s *CatalogService) GetProduct(ctx context.Context, req *travel.ProductIdRequest) (*travel.Product, error) {
	if req == nil || req.GetId() <= 0 { return nil, status.Error(codes.InvalidArgument, "product id is required") }
	tenantID, err := auth.TenantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	item, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil { if ent.IsNotFound(err) { return nil, status.Error(codes.NotFound, "product not found") }; return nil, status.Error(codes.Internal, err.Error()) }
	if item.TenantID != tenantID || item.Status != "ACTIVE" { return nil, status.Error(codes.NotFound, "product not found") }
	return toProduct(item), nil
}

func (s *CatalogService) ListProducts(ctx context.Context, req *travel.ProductListRequest) (*travel.ProductListResponse, error) {
	if req == nil { req = &travel.ProductListRequest{} }
	tenantID, err := auth.TenantID(ctx); if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	page, pageSize := int(req.GetPage()), int(req.GetPageSize()); if page <= 0 { page = 1 }; if pageSize <= 0 { pageSize = 20 }; if pageSize > 100 { pageSize = 100 }
	items, total, err := s.repo.List(ctx, tenantID, req.GetKeyword(), req.GetDestination(), (page-1)*pageSize, pageSize)
	if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
	result := &travel.ProductListResponse{Total:int64(total)}
	for _, item := range items { if item.Status == "ACTIVE" { result.Items = append(result.Items, toProduct(item)) } }
	return result, nil
}

func toProduct(item *ent.Product) *travel.Product {
	return &travel.Product{Id:item.ID, TenantId:item.TenantID, MerchantId:item.MerchantID, Code:item.Code, Title:item.Title, Slug:item.Slug, Destination:item.Destination, Description:item.Description, Currency:item.Currency, MinPrice:item.MinPrice, Status:item.Status}
}
