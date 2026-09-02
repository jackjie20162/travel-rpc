package service

import (
    "context"
    "strings"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/ent/inventory"
    "gitee.com/meinongyihe/travel-rpc/ent/product"
    "gitee.com/meinongyihe/travel-rpc/ent/productpackage"
    "gitee.com/meinongyihe/travel-rpc/internal/auth"
    "gitee.com/meinongyihe/travel-rpc/travel"
)

// ManagementService owns merchant-side catalog publication and inventory setup.
// Tenant and merchant scope always comes from authenticated RPC metadata.
type ManagementService struct {
    travel.UnimplementedTravelManagementServiceServer
    client *ent.Client
}

func NewManagementService(client *ent.Client) *ManagementService { return &ManagementService{client: client} }

func scope(ctx context.Context) (int64, int64, error) {
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return 0, 0, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return 0, 0, status.Error(codes.Unauthenticated, err.Error()) }
    return tenantID, merchantID, nil
}

func (s *ManagementService) CreateProduct(ctx context.Context, req *travel.CreateProductRequest) (*travel.Product, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    if strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetTitle()) == "" { return nil, status.Error(codes.InvalidArgument, "code and title are required") }
    currency := strings.ToUpper(strings.TrimSpace(req.GetCurrency())); if currency == "" { currency = "AED" }
    p, err := s.client.Product.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetCode(strings.TrimSpace(req.GetCode())).SetTitle(strings.TrimSpace(req.GetTitle())).SetSlug(strings.TrimSpace(req.GetSlug())).SetDestination(strings.TrimSpace(req.GetDestination())).SetDescription(req.GetDescription()).SetCurrency(currency).SetMinPrice(req.GetMinPrice()).SetStatus("DRAFT").Save(ctx)
    if err != nil { return nil, status.Error(codes.AlreadyExists, err.Error()) }
    return productMessage(p), nil
}

func (s *ManagementService) UpdateProduct(ctx context.Context, req *travel.UpdateProductRequest) (*travel.Product, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    if req.GetId() <= 0 || strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetTitle()) == "" { return nil, status.Error(codes.InvalidArgument, "id, code and title are required") }
    p, err := s.client.Product.Query().Where(product.IDEQ(int(req.GetId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
    if err != nil { if ent.IsNotFound(err) { return nil, status.Error(codes.NotFound, "product not found") }; return nil, err }
    if p.Status == "PUBLISHED" { return nil, status.Error(codes.FailedPrecondition, "published product must be unpublished before editing") }
    p, err = p.Update().SetCode(strings.TrimSpace(req.GetCode())).SetTitle(strings.TrimSpace(req.GetTitle())).SetSlug(strings.TrimSpace(req.GetSlug())).SetDestination(strings.TrimSpace(req.GetDestination())).SetDescription(req.GetDescription()).SetCurrency(strings.ToUpper(strings.TrimSpace(req.GetCurrency()))).SetMinPrice(req.GetMinPrice()).Save(ctx)
    if err != nil { return nil, status.Error(codes.AlreadyExists, err.Error()) }
    return productMessage(p), nil
}

func (s *ManagementService) CreatePackage(ctx context.Context, req *travel.CreatePackageRequest) (*travel.ProductPackage, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    if req.GetProductId() <= 0 || strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetName()) == "" { return nil, status.Error(codes.InvalidArgument, "productId, code and name are required") }
    p, err := s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
    if err != nil { if ent.IsNotFound(err) { return nil, status.Error(codes.NotFound, "product not found") }; return nil, err }
    if p.Status == "PUBLISHED" { return nil, status.Error(codes.FailedPrecondition, "published product cannot add packages") }
    pkg, err := s.client.ProductPackage.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetProductID(p.ID).SetCode(strings.TrimSpace(req.GetCode())).SetName(strings.TrimSpace(req.GetName())).SetStatus("ACTIVE").Save(ctx)
    if err != nil { return nil, status.Error(codes.AlreadyExists, err.Error()) }
    return packageMessage(pkg), nil
}

func (s *ManagementService) ListPackages(ctx context.Context, req *travel.PackageListRequest) (*travel.PackageListResponse, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    items, err := s.client.ProductPackage.Query().Where(productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID), productpackage.ProductIDEQ(int(req.GetProductId()))).Order(ent.Asc(productpackage.FieldID)).All(ctx)
    if err != nil { return nil, err }
    out := &travel.PackageListResponse{}
    for _, item := range items { out.Items = append(out.Items, packageMessage(item)) }
    return out, nil
}

func (s *ManagementService) UpsertInventory(ctx context.Context, req *travel.UpsertInventoryRequest) (*travel.InventoryItem, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    if req.GetPackageId() <= 0 || strings.TrimSpace(req.GetDate()) == "" || req.GetCapacity() < 0 || req.GetUnitPrice() < 0 { return nil, status.Error(codes.InvalidArgument, "packageId, date, capacity and unitPrice are required") }
    pkg, err := s.client.ProductPackage.Query().Where(productpackage.IDEQ(int(req.GetPackageId())), productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID)).Only(ctx)
    if err != nil { if ent.IsNotFound(err) { return nil, status.Error(codes.NotFound, "package not found") }; return nil, err }
    timeSlot := strings.TrimSpace(req.GetTimeSlot())
    currency := strings.ToUpper(strings.TrimSpace(req.GetCurrency())); if currency == "" { currency = "AED" }
    statusValue := strings.ToUpper(strings.TrimSpace(req.GetStatus())); if statusValue == "" { statusValue = "OPEN" }
    q := s.client.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.MerchantIDEQ(merchantID), inventory.PackageIDEQ(pkg.ID), inventory.ServiceDateEQ(req.GetDate()), inventory.TimeSlotEQ(timeSlot))
    item, err := q.Only(ctx)
    if ent.IsNotFound(err) {
        item, err = s.client.Inventory.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetPackageID(pkg.ID).SetServiceDate(req.GetDate()).SetTimeSlot(timeSlot).SetCapacity(int(req.GetCapacity())).SetReserved(0).SetUnitPrice(req.GetUnitPrice()).SetCurrency(currency).SetStatus(statusValue).Save(ctx)
    } else if err == nil {
        if item.Reserved > int(req.GetCapacity()) { return nil, status.Error(codes.FailedPrecondition, "capacity cannot be below reserved quantity") }
        item, err = item.Update().SetCapacity(int(req.GetCapacity())).SetUnitPrice(req.GetUnitPrice()).SetCurrency(currency).SetStatus(statusValue).Save(ctx)
    }
    if err != nil { return nil, err }
    return inventoryMessage(item), nil
}

func (s *ManagementService) ListInventory(ctx context.Context, req *travel.InventoryListRequest) (*travel.InventoryListResponse, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    items, err := s.client.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.MerchantIDEQ(merchantID), inventory.PackageIDEQ(int(req.GetPackageId()))).Order(ent.Asc(inventory.FieldServiceDate), ent.Asc(inventory.FieldTimeSlot)).All(ctx)
    if err != nil { return nil, err }
    out := &travel.InventoryListResponse{}
    for _, item := range items { out.Items = append(out.Items, inventoryMessage(item)) }
    return out, nil
}

func (s *ManagementService) PublishProduct(ctx context.Context, req *travel.PublishProductRequest) (*travel.Product, error) {
    tenantID, merchantID, err := scope(ctx); if err != nil { return nil, err }
    p, err := s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
    if err != nil { if ent.IsNotFound(err) { return nil, status.Error(codes.NotFound, "product not found") }; return nil, err }
    target := "DRAFT"; if req.GetPublished() { target = "PUBLISHED" }
    if req.GetPublished() {
        count, err := s.client.ProductPackage.Query().Where(productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID), productpackage.ProductIDEQ(p.ID), productpackage.StatusEQ("ACTIVE")).Count(ctx)
        if err != nil { return nil, err }; if count == 0 { return nil, status.Error(codes.FailedPrecondition, "product requires at least one active package") }
    }
    p, err = p.Update().SetStatus(target).Save(ctx); if err != nil { return nil, err }
    return productMessage(p), nil
}

func productMessage(p *ent.Product) *travel.Product { return &travel.Product{Id:int64(p.ID), TenantId:p.TenantID, MerchantId:p.MerchantID, Code:p.Code, Title:p.Title, Slug:p.Slug, Destination:p.Destination, Description:p.Description, Currency:p.Currency, MinPrice:p.MinPrice, Status:p.Status} }
func packageMessage(p *ent.ProductPackage) *travel.ProductPackage { return &travel.ProductPackage{Id:int64(p.ID), ProductId:p.ProductID, Code:p.Code, Name:p.Name, Status:p.Status} }
func inventoryMessage(i *ent.Inventory) *travel.InventoryItem { return &travel.InventoryItem{Id:int64(i.ID), PackageId:i.PackageID, Date:i.ServiceDate, TimeSlot:i.TimeSlot, Capacity:int32(i.Capacity), Reserved:int32(i.Reserved), UnitPrice:i.UnitPrice, Currency:i.Currency, Status:i.Status} }
