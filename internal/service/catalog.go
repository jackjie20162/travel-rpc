package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/productpackage"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService struct {
	travel.UnimplementedCatalogServiceServer
	repo   repository.ProductRepository
	client *ent.Client
}

func NewCatalogService(repo repository.ProductRepository, client *ent.Client) *CatalogService {
	return &CatalogService{repo: repo, client: client}
}

func (s *CatalogService) GetProduct(ctx context.Context, req *travel.ProductIdRequest) (*travel.Product, error) {
	if req == nil || req.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product id is required")
	}
	item, err := s.repo.GetByID(ctx, req.GetId())
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return toProduct(item), nil
}

func (s *CatalogService) ListProducts(ctx context.Context, req *travel.ProductListRequest) (*travel.ProductListResponse, error) {
	if req == nil {
		req = &travel.ProductListRequest{}
	}
	tenantID, _ := auth.TenantID(ctx) // optional for public catalog
	page, pageSize := int(req.GetPage()), int(req.GetPageSize())
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}
	items, total, err := s.repo.List(ctx, tenantID, req.GetKeyword(), req.GetDestination(), (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := &travel.ProductListResponse{Total: int64(total)}
	for _, item := range items {
		if item.Status == "PUBLISHED" {
			result.Items = append(result.Items, toProduct(item))
		}
	}
	return result, nil
}

func toProduct(item *ent.Product) *travel.Product {
	return &travel.Product{Id: int64(item.ID), TenantId: item.TenantID, MerchantId: item.MerchantID, Code: item.Code, Title: item.Title, Slug: item.Slug, Destination: item.Destination, Description: item.Description, Currency: item.Currency, MinPrice: item.MinPrice, Status: item.Status}
}

func (s *CatalogService) ListPackages(ctx context.Context, req *travel.PackageListRequest) (*travel.PackageListResponse, error) {
	if req == nil || req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	tenantID, _ := auth.TenantID(ctx) // optional for public catalog
	q := s.client.ProductPackage.Query().Where(productpackage.ProductIDEQ(req.GetProductId()))
	if tenantID > 0 {
		q = q.Where(productpackage.TenantIDEQ(tenantID))
	}
	items, err := q.Order(ent.Asc(productpackage.FieldID)).All(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := &travel.PackageListResponse{}
	for _, item := range items {
		out.Items = append(out.Items, packageMessage(item))
	}
	return out, nil
}
