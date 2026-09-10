package service

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/itinerarystop"
	"gitee.com/meinongyihe/travel-rpc/ent/productpackage"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/internal/repository"
	"gitee.com/meinongyihe/travel-rpc/travel"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type CatalogService struct {
	travel.UnimplementedCatalogServiceServer
	repo         repository.ProductRepository
	client       *ent.Client
	translations repository.TranslationRepository
}

func NewCatalogService(repo repository.ProductRepository, client *ent.Client, translations repository.TranslationRepository) *CatalogService {
	return &CatalogService{repo: repo, client: client, translations: translations}
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
	p := productMessage(item)
	s.localizeProducts(ctx, []*travel.Product{p})
	return p, nil
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
			result.Items = append(result.Items, productMessage(item))
		}
	}
	s.localizeProducts(ctx, result.Items)
	return result, nil
}

// localizeProducts overrides base-language text with the request-locale
// translation when a DONE row exists. The lookup is batched for the whole page.
func (s *CatalogService) localizeProducts(ctx context.Context, products []*travel.Product) {
	if s.translations == nil || len(products) == 0 {
		return
	}
	locale := auth.Locale(ctx)
	if locale == "" {
		return
	}
	ids := make([]int64, 0, len(products))
	for _, p := range products {
		ids = append(ids, p.GetId())
	}
	m, err := s.translations.MapProducts(ctx, ids, locale)
	if err != nil {
		logx.Errorf("[catalog] load product translations (locale=%s): %v", locale, err)
		return
	}
	for _, p := range products {
		applyProductTranslation(p, m[p.GetId()])
	}
}

// applyProductTranslation overrides translatable text fields when a translation
// exists; empty translation fields keep the base-language value.
func applyProductTranslation(p *travel.Product, t *ent.ProductTranslation) {
	if p == nil || t == nil {
		return
	}
	if t.Title != "" {
		p.Title = t.Title
	}
	if t.Description != "" {
		p.Description = t.Description
	}
	if t.Highlights != "" {
		p.Highlights = t.Highlights
	}
	if t.RichContent != "" {
		p.RichContent = t.RichContent
	}
	if t.BookingNotice != "" {
		p.BookingNotice = t.BookingNotice
	}
}

// localizePackages overrides package names with the request-locale translation.
func (s *CatalogService) localizePackages(ctx context.Context, packages []*travel.ProductPackage) {
	if s.translations == nil || len(packages) == 0 {
		return
	}
	locale := auth.Locale(ctx)
	if locale == "" {
		return
	}
	ids := make([]int64, 0, len(packages))
	for _, p := range packages {
		ids = append(ids, p.GetId())
	}
	m, err := s.translations.MapPackages(ctx, ids, locale)
	if err != nil {
		logx.Errorf("[catalog] load package translations (locale=%s): %v", locale, err)
		return
	}
	for _, p := range packages {
		if t := m[p.GetId()]; t != nil && t.Name != "" {
			p.Name = t.Name
		}
	}
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
	s.localizePackages(ctx, out.Items)
	return out, nil
}

// ListItineraryStops C 端公开读取产品行程节点（租户可选，按 sequence 排序）
func (s *CatalogService) ListItineraryStops(ctx context.Context, req *travel.ItineraryStopListRequest) (*travel.ItineraryStopListResponse, error) {
	if req == nil || req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "product_id is required")
	}
	tenantID, _ := auth.TenantID(ctx) // optional for public catalog
	q := s.client.ItineraryStop.Query().Where(itinerarystop.ProductIDEQ(req.GetProductId()))
	if tenantID > 0 {
		q = q.Where(itinerarystop.TenantIDEQ(tenantID))
	}
	items, err := q.Order(ent.Asc(itinerarystop.FieldSequence), ent.Asc(itinerarystop.FieldID)).All(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := &travel.ItineraryStopListResponse{}
	for _, item := range items {
		out.Items = append(out.Items, itineraryStopMessage(item))
	}
	return out, nil
}
