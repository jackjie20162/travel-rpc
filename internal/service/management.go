package service

import (
	"context"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/inventory"
	"gitee.com/meinongyihe/travel-rpc/ent/itinerarystop"
	"gitee.com/meinongyihe/travel-rpc/ent/product"
	"gitee.com/meinongyihe/travel-rpc/ent/productpackage"
	"gitee.com/meinongyihe/travel-rpc/internal/auth"
	"gitee.com/meinongyihe/travel-rpc/travel"

	"github.com/zeromicro/go-zero/core/logx"
)

// ManagementService owns merchant-side catalog publication and inventory setup.
// Tenant and merchant scope always comes from authenticated RPC metadata.
type ManagementService struct {
	travel.UnimplementedTravelManagementServiceServer
	client *ent.Client
}

func NewManagementService(client *ent.Client) *ManagementService {
	return &ManagementService{client: client}
}

// withTx 在事务中执行 fn，自动处理 Commit/Rollback，避免写漏
func (s *ManagementService) withTx(ctx context.Context, fn func(tx *ent.Tx) error) error {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			_ = tx.Rollback()
			panic(v)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (s *ManagementService) ListProducts(ctx context.Context, req *travel.ProductListRequest) (*travel.ProductListResponse, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req == nil {
		req = &travel.ProductListRequest{}
	}
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
	q := s.client.Product.Query().Where(product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID))
	if kw := req.GetKeyword(); kw != "" {
		q = q.Where(product.TitleContains(kw))
	}
	if dst := req.GetDestination(); dst != "" {
		q = q.Where(product.DestinationEQ(dst))
	}
	total, err := q.Count(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	items, err := q.Order(ent.Desc(product.FieldID)).Offset((page - 1) * pageSize).Limit(pageSize).All(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	result := &travel.ProductListResponse{Total: int64(total)}
	for _, item := range items {
		result.Items = append(result.Items, productMessage(item))
	}
	return result, nil
}

func scope(ctx context.Context) (int64, int64, error) {
	tenantID, err := auth.TenantID(ctx)
	if err != nil {
		return 0, 0, status.Error(codes.Unauthenticated, err.Error())
	}
	merchantID, err := auth.MerchantID(ctx)
	if err != nil {
		return 0, 0, status.Error(codes.Unauthenticated, err.Error())
	}
	return tenantID, merchantID, nil
}

func (s *ManagementService) CreateProduct(ctx context.Context, req *travel.CreateProductRequest) (*travel.Product, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetTitle()) == "" {
		return nil, status.Error(codes.InvalidArgument, "code and title are required")
	}
	currency := strings.ToUpper(strings.TrimSpace(req.GetCurrency()))
	if currency == "" {
		currency = "AED"
	}
	create := s.client.Product.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetCode(strings.TrimSpace(req.GetCode())).SetTitle(strings.TrimSpace(req.GetTitle())).SetSlug(strings.TrimSpace(req.GetSlug())).SetDestination(strings.TrimSpace(req.GetDestination())).SetDescription(req.GetDescription()).SetCurrency(currency).SetMinPrice(req.GetMinPrice()).SetStatus("DRAFT")
	if req.GetHighlights() != "" {
		create = create.SetHighlights(req.GetHighlights())
	}
	if req.GetCoverImage() != "" {
		create = create.SetCoverImage(req.GetCoverImage())
	}
	if req.GetImages() != "" {
		create = create.SetImages(req.GetImages())
	}
	if req.GetVideoUrl() != "" {
		create = create.SetVideoURL(req.GetVideoUrl())
	}
	if req.GetRichContent() != "" {
		create = create.SetRichContent(req.GetRichContent())
	}
	if req.GetBookingNotice() != "" {
		create = create.SetBookingNotice(req.GetBookingNotice())
	}
	p, err := create.Save(ctx)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return productMessage(p), nil
}

func (s *ManagementService) UpdateProduct(ctx context.Context, req *travel.UpdateProductRequest) (*travel.Product, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() <= 0 || strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetTitle()) == "" {
		return nil, status.Error(codes.InvalidArgument, "id, code and title are required")
	}

	var result *travel.Product
	err = s.withTx(ctx, func(tx *ent.Tx) error {
		// 在事务中查询产品
		p, err := tx.Product.Query().Where(product.IDEQ(int(req.GetId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return status.Error(codes.NotFound, "product not found")
			}
			return err
		}

		newCurrency := strings.ToUpper(strings.TrimSpace(req.GetCurrency()))
		oldCurrency := p.Currency

		// 构建产品更新
		update := p.Update().SetCode(strings.TrimSpace(req.GetCode())).SetTitle(strings.TrimSpace(req.GetTitle())).SetSlug(strings.TrimSpace(req.GetSlug())).SetDestination(strings.TrimSpace(req.GetDestination())).SetDescription(req.GetDescription()).SetCurrency(newCurrency).SetMinPrice(req.GetMinPrice())
		if req.GetHighlights() != "" {
			update = update.SetHighlights(req.GetHighlights())
		} else {
			update = update.ClearHighlights()
		}
		if req.GetCoverImage() != "" {
			update = update.SetCoverImage(req.GetCoverImage())
		} else {
			update = update.ClearCoverImage()
		}
		if req.GetImages() != "" {
			update = update.SetImages(req.GetImages())
		} else {
			update = update.ClearImages()
		}
		if req.GetVideoUrl() != "" {
			update = update.SetVideoURL(req.GetVideoUrl())
		} else {
			update = update.ClearVideoURL()
		}
		if req.GetRichContent() != "" {
			update = update.SetRichContent(req.GetRichContent())
		} else {
			update = update.ClearRichContent()
		}
		if req.GetBookingNotice() != "" {
			update = update.SetBookingNotice(req.GetBookingNotice())
		} else {
			update = update.ClearBookingNotice()
		}

		// 在事务中保存产品
		p, err = update.Save(ctx)
		if err != nil {
			return err
		}

		// 币种变更时，在同一个事务中级联更新套餐和库存的币种
		if newCurrency != "" && newCurrency != oldCurrency {
			logx.Infof("[UpdateProduct] currency changed: %s -> %s, productID=%d, tenantID=%d, merchantID=%d", oldCurrency, newCurrency, p.ID, tenantID, merchantID)

			// 更新所有关联套餐的卖价/底价币种
			pkgAffected, err := tx.ProductPackage.Update().
				Where(
					productpackage.TenantIDEQ(tenantID),
					productpackage.MerchantIDEQ(merchantID),
					productpackage.ProductIDEQ(int64(p.ID)),
				).
				SetSellCurrency(newCurrency).
				SetCostCurrency(newCurrency).
				Save(ctx)
			if err != nil {
				return err
			}
			logx.Infof("[UpdateProduct] updated %d packages currency", pkgAffected)

			// 查询关联套餐，获取套餐 ID 列表
			pkgs, err := tx.ProductPackage.Query().
				Where(
					productpackage.TenantIDEQ(tenantID),
					productpackage.MerchantIDEQ(merchantID),
					productpackage.ProductIDEQ(int64(p.ID)),
				).
				All(ctx)
			if err != nil {
				return err
			}
			logx.Infof("[UpdateProduct] found %d packages for product %d", len(pkgs), p.ID)

			if len(pkgs) > 0 {
				pkgID64s := make([]int64, len(pkgs))
				for i, pkg := range pkgs {
					pkgID64s[i] = int64(pkg.ID)
				}
				invAffected, err := tx.Inventory.Update().
					Where(
						inventory.TenantIDEQ(tenantID),
						inventory.MerchantIDEQ(merchantID),
						inventory.PackageIDIn(pkgID64s...),
					).
					SetCurrency(newCurrency).
					Save(ctx)
				if err != nil {
					return err
				}
				logx.Infof("[UpdateProduct] updated %d inventories currency", invAffected)
			}
		}

		result = productMessage(p)
		return nil
	})

	if err != nil {
		if se, ok := status.FromError(err); ok {
			return nil, se.Err()
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return result, nil
}

func (s *ManagementService) CreatePackage(ctx context.Context, req *travel.CreatePackageRequest) (*travel.ProductPackage, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetProductId() <= 0 || strings.TrimSpace(req.GetCode()) == "" || strings.TrimSpace(req.GetName()) == "" {
		return nil, status.Error(codes.InvalidArgument, "productId, code and name are required")
	}
	p, err := s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, err
	}
	if p.Status == "PUBLISHED" {
		return nil, status.Error(codes.FailedPrecondition, "published product cannot add packages")
	}
	pricingMode := strings.TrimSpace(req.GetPricingMode())
	if pricingMode == "" {
		pricingMode = "SAME_PRICE"
	}
	inventoryMode := strings.TrimSpace(req.GetInventoryMode())
	if inventoryMode == "" {
		inventoryMode = "UNLIMITED"
	}
	sellCurrency := strings.ToUpper(strings.TrimSpace(req.GetSellCurrency()))
	if sellCurrency == "" {
		sellCurrency = "AED"
	}
	costCurrency := strings.ToUpper(strings.TrimSpace(req.GetCostCurrency()))
	if costCurrency == "" {
		costCurrency = "AED"
	}
	minOrderQty := int(req.GetMinOrderQty())
	if minOrderQty <= 0 {
		minOrderQty = 1
	}
	create := s.client.ProductPackage.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetProductID(int64(p.ID)).SetCode(strings.TrimSpace(req.GetCode())).SetName(strings.TrimSpace(req.GetName())).SetStatus("ACTIVE").SetPricingMode(pricingMode).SetInventoryMode(inventoryMode).SetMinOrderQty(minOrderQty).SetSellCurrency(sellCurrency).SetCostCurrency(costCurrency)
	if req.GetGroupPrices() != "" {
		create = create.SetGroupPrices(req.GetGroupPrices())
	}
	if req.GetTierPrices() != "" {
		create = create.SetTierPrices(req.GetTierPrices())
	}
	pkg, err := create.Save(ctx)
	if err != nil {
		return nil, status.Error(codes.AlreadyExists, err.Error())
	}
	return packageMessage(pkg), nil
}

func (s *ManagementService) ListPackages(ctx context.Context, req *travel.PackageListRequest) (*travel.PackageListResponse, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.client.ProductPackage.Query().Where(productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID), productpackage.ProductIDEQ(req.GetProductId())).Order(ent.Asc(productpackage.FieldID)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := &travel.PackageListResponse{}
	for _, item := range items {
		out.Items = append(out.Items, packageMessage(item))
	}
	return out, nil
}

func (s *ManagementService) UpsertInventory(ctx context.Context, req *travel.UpsertInventoryRequest) (*travel.InventoryItem, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetPackageId() <= 0 || strings.TrimSpace(req.GetDate()) == "" || req.GetCapacity() < 0 || req.GetUnitPrice() < 0 {
		return nil, status.Error(codes.InvalidArgument, "packageId, date, capacity and unitPrice are required")
	}
	pkg, err := s.client.ProductPackage.Query().Where(productpackage.IDEQ(int(req.GetPackageId())), productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "package not found")
		}
		return nil, err
	}
	timeSlot := strings.TrimSpace(req.GetTimeSlot())
	currency := strings.ToUpper(strings.TrimSpace(req.GetCurrency()))
	if currency == "" {
		currency = "AED"
	}
	statusValue := strings.ToUpper(strings.TrimSpace(req.GetStatus()))
	if statusValue == "" {
		statusValue = "OPEN"
	}
	inventoryMode := strings.TrimSpace(req.GetInventoryMode())
	if inventoryMode == "" {
		inventoryMode = "UNLIMITED"
	}
	q := s.client.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.MerchantIDEQ(merchantID), inventory.PackageIDEQ(int64(pkg.ID)), inventory.ServiceDateEQ(req.GetDate()), inventory.TimeSlotEQ(timeSlot))
	item, err := q.Only(ctx)
	if ent.IsNotFound(err) {
		create := s.client.Inventory.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetPackageID(int64(pkg.ID)).SetServiceDate(req.GetDate()).SetTimeSlot(timeSlot).SetCapacity(int(req.GetCapacity())).SetReserved(0).SetUnitPrice(req.GetUnitPrice()).SetCurrency(currency).SetStatus(statusValue).SetInventoryMode(inventoryMode).SetIsOpen(true)
		if req.GetTotalCapacity() > 0 {
			create = create.SetTotalCapacity(int(req.GetTotalCapacity()))
		}
		item, err = create.Save(ctx)
	} else if err == nil {
		if item.Reserved > int(req.GetCapacity()) {
			return nil, status.Error(codes.FailedPrecondition, "capacity cannot be below reserved quantity")
		}
		update := item.Update().SetCapacity(int(req.GetCapacity())).SetUnitPrice(req.GetUnitPrice()).SetCurrency(currency).SetStatus(statusValue).SetInventoryMode(inventoryMode).SetIsOpen(req.GetIsOpen())
		if req.GetTotalCapacity() > 0 {
			update = update.SetTotalCapacity(int(req.GetTotalCapacity()))
		}
		item, err = update.Save(ctx)
	}
	if err != nil {
		return nil, err
	}
	return inventoryMessage(item), nil
}

func (s *ManagementService) ListInventory(ctx context.Context, req *travel.InventoryListRequest) (*travel.InventoryListResponse, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	items, err := s.client.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.MerchantIDEQ(merchantID), inventory.PackageIDEQ(req.GetPackageId())).Order(ent.Asc(inventory.FieldServiceDate), ent.Asc(inventory.FieldTimeSlot)).All(ctx)
	if err != nil {
		return nil, err
	}
	out := &travel.InventoryListResponse{}
	for _, item := range items {
		out.Items = append(out.Items, inventoryMessage(item))
	}
	return out, nil
}

func (s *ManagementService) PublishProduct(ctx context.Context, req *travel.PublishProductRequest) (*travel.Product, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	p, err := s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, err
	}
	target := "DRAFT"
	if req.GetPublished() {
		target = "PUBLISHED"
	}
	if req.GetPublished() {
		count, err := s.client.ProductPackage.Query().Where(productpackage.TenantIDEQ(tenantID), productpackage.MerchantIDEQ(merchantID), productpackage.ProductIDEQ(int64(p.ID)), productpackage.StatusEQ("ACTIVE")).Count(ctx)
		if err != nil {
			return nil, err
		}
		if count == 0 {
			return nil, status.Error(codes.FailedPrecondition, "product requires at least one active package")
		}
	}
	p, err = p.Update().SetStatus(target).Save(ctx)
	if err != nil {
		return nil, err
	}
	return productMessage(p), nil
}

func productMessage(p *ent.Product) *travel.Product {
	return &travel.Product{
		Id: int64(p.ID), TenantId: p.TenantID, MerchantId: p.MerchantID,
		Code: p.Code, Title: p.Title, Slug: p.Slug, Destination: p.Destination,
		Description: p.Description, Currency: p.Currency, MinPrice: p.MinPrice, Status: p.Status,
		Highlights: p.Highlights, CoverImage: p.CoverImage, Images: p.Images,
		VideoUrl: p.VideoURL, RichContent: p.RichContent, BookingNotice: p.BookingNotice,
	}
}

func packageMessage(p *ent.ProductPackage) *travel.ProductPackage {
	return &travel.ProductPackage{
		Id: int64(p.ID), ProductId: p.ProductID, Code: p.Code, Name: p.Name, Status: p.Status,
		PricingMode: p.PricingMode, InventoryMode: p.InventoryMode,
		MinOrderQty: int32(p.MinOrderQty), SellCurrency: p.SellCurrency, CostCurrency: p.CostCurrency,
		GroupPrices: p.GroupPrices, TierPrices: p.TierPrices,
	}
}

func inventoryMessage(i *ent.Inventory) *travel.InventoryItem {
	return &travel.InventoryItem{
		Id: int64(i.ID), PackageId: i.PackageID, Date: i.ServiceDate, TimeSlot: i.TimeSlot,
		Capacity: int32(i.Capacity), Reserved: int32(i.Reserved), UnitPrice: i.UnitPrice,
		Currency: i.Currency, Status: i.Status, InventoryMode: i.InventoryMode,
		TotalCapacity: int32(i.TotalCapacity), IsOpen: i.IsOpen,
	}
}

func itineraryStopMessage(s *ent.ItineraryStop) *travel.ItineraryStop {
	return &travel.ItineraryStop{
		Id: int64(s.ID), ProductId: s.ProductID, StopType: s.StopType, Title: s.Title,
		Description: s.Description, Sequence: int32(s.Sequence), LocationName: s.LocationName,
		Address: s.Address, Latitude: s.Latitude, Longitude: s.Longitude,
		DurationMinutes: int32(s.DurationMinutes), TransportType: s.TransportType,
		StartTime: s.StartTime, EndTime: s.EndTime,
		IsPickup: s.IsPickup, IsDropoff: s.IsDropoff,
		AgreementNoShopping: s.AgreementNoShopping, AgreementAdjustable: s.AgreementAdjustable,
		PoiId: s.PoiID, PoiName: s.PoiName,
		IsEntering: s.IsEntering, DurationMode: s.DurationMode,
		DurationHours: int32(s.DurationHours), ActivityFeatures: s.ActivityFeatures,
		PickupLocation: s.PickupLocation, PickupAddress: s.PickupAddress,
		PickupLatitude: s.PickupLatitude, PickupLongitude: s.PickupLongitude,
		DropoffLocation: s.DropoffLocation, DropoffAddress: s.DropoffAddress,
		DropoffLatitude: s.DropoffLatitude, DropoffLongitude: s.DropoffLongitude,
		TypeParams: s.TypeParams,
	}
}

// Itinerary Stop CRUD methods

func (s *ManagementService) CreateItineraryStop(ctx context.Context, req *travel.CreateItineraryStopRequest) (*travel.ItineraryStop, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetProductId() <= 0 || strings.TrimSpace(req.GetTitle()) == "" {
		return nil, status.Error(codes.InvalidArgument, "productId and title are required")
	}
	// 已发布产品也允许修改行程（仅校验产品归属）
	_, err = s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, err
	}
	stopType := strings.TrimSpace(req.GetStopType())
	if stopType == "" {
		stopType = "ACTIVITY"
	}
	create := s.client.ItineraryStop.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetProductID(req.GetProductId()).SetStopType(stopType).SetTitle(strings.TrimSpace(req.GetTitle())).SetSequence(int(req.GetSequence()))
	if req.GetDescription() != "" {
		create = create.SetDescription(req.GetDescription())
	}
	if req.GetLocationName() != "" {
		create = create.SetLocationName(req.GetLocationName())
	}
	if req.GetAddress() != "" {
		create = create.SetAddress(req.GetAddress())
	}
	if req.GetLatitude() != 0 {
		create = create.SetLatitude(req.GetLatitude())
	}
	if req.GetLongitude() != 0 {
		create = create.SetLongitude(req.GetLongitude())
	}
	if req.GetDurationMinutes() > 0 {
		create = create.SetDurationMinutes(int(req.GetDurationMinutes()))
	}
	if req.GetTransportType() != "" {
		create = create.SetTransportType(req.GetTransportType())
	}
	if req.GetStartTime() != "" {
		create = create.SetStartTime(req.GetStartTime())
	}
	if req.GetEndTime() != "" {
		create = create.SetEndTime(req.GetEndTime())
	}
	if req.GetIsPickup() {
		create = create.SetIsPickup(true)
	}
	if req.GetIsDropoff() {
		create = create.SetIsDropoff(true)
	}
	if req.GetAgreementNoShopping() {
		create = create.SetAgreementNoShopping(true)
	}
	if req.GetAgreementAdjustable() {
		create = create.SetAgreementAdjustable(true)
	}
	// 新增字段
	if req.GetPoiId() != "" {
		create = create.SetPoiID(req.GetPoiId())
	}
	if req.GetPoiName() != "" {
		create = create.SetPoiName(req.GetPoiName())
	}
	create = create.SetIsEntering(req.GetIsEntering())
	if req.GetDurationMode() != "" {
		create = create.SetDurationMode(req.GetDurationMode())
	}
	if req.GetDurationHours() > 0 {
		create = create.SetDurationHours(int(req.GetDurationHours()))
	}
	if req.GetActivityFeatures() != "" {
		create = create.SetActivityFeatures(req.GetActivityFeatures())
	}
	if req.GetPickupLocation() != "" {
		create = create.SetPickupLocation(req.GetPickupLocation())
	}
	if req.GetPickupAddress() != "" {
		create = create.SetPickupAddress(req.GetPickupAddress())
	}
	if req.GetPickupLatitude() != 0 {
		create = create.SetPickupLatitude(req.GetPickupLatitude())
	}
	if req.GetPickupLongitude() != 0 {
		create = create.SetPickupLongitude(req.GetPickupLongitude())
	}
	if req.GetDropoffLocation() != "" {
		create = create.SetDropoffLocation(req.GetDropoffLocation())
	}
	if req.GetDropoffAddress() != "" {
		create = create.SetDropoffAddress(req.GetDropoffAddress())
	}
	if req.GetDropoffLatitude() != 0 {
		create = create.SetDropoffLatitude(req.GetDropoffLatitude())
	}
	if req.GetDropoffLongitude() != 0 {
		create = create.SetDropoffLongitude(req.GetDropoffLongitude())
	}
	if req.GetTypeParams() != "" {
		create = create.SetTypeParams(req.GetTypeParams())
	}
	stop, err := create.Save(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return itineraryStopMessage(stop), nil
}

func (s *ManagementService) UpdateItineraryStop(ctx context.Context, req *travel.UpdateItineraryStopRequest) (*travel.ItineraryStop, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() <= 0 || req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id and productId are required")
	}
	stop, err := s.client.ItineraryStop.Query().Where(itinerarystop.IDEQ(int(req.GetId())), itinerarystop.ProductIDEQ(req.GetProductId()), itinerarystop.TenantIDEQ(tenantID), itinerarystop.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "itinerary stop not found")
		}
		return nil, err
	}
	_, err = s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, err
	}
	stopType := strings.TrimSpace(req.GetStopType())
	if stopType == "" {
		stopType = stop.StopType
	}
	update := stop.Update().SetStopType(stopType).SetTitle(strings.TrimSpace(req.GetTitle())).SetSequence(int(req.GetSequence()))
	if req.GetDescription() != "" {
		update = update.SetDescription(req.GetDescription())
	} else {
		update = update.ClearDescription()
	}
	if req.GetLocationName() != "" {
		update = update.SetLocationName(req.GetLocationName())
	} else {
		update = update.ClearLocationName()
	}
	if req.GetAddress() != "" {
		update = update.SetAddress(req.GetAddress())
	} else {
		update = update.ClearAddress()
	}
	if req.GetLatitude() != 0 {
		update = update.SetLatitude(req.GetLatitude())
	} else {
		update = update.ClearLatitude()
	}
	if req.GetLongitude() != 0 {
		update = update.SetLongitude(req.GetLongitude())
	} else {
		update = update.ClearLongitude()
	}
	if req.GetDurationMinutes() > 0 {
		update = update.SetDurationMinutes(int(req.GetDurationMinutes()))
	} else {
		update = update.ClearDurationMinutes()
	}
	if req.GetTransportType() != "" {
		update = update.SetTransportType(req.GetTransportType())
	} else {
		update = update.ClearTransportType()
	}
	if req.GetStartTime() != "" {
		update = update.SetStartTime(req.GetStartTime())
	} else {
		update = update.ClearStartTime()
	}
	if req.GetEndTime() != "" {
		update = update.SetEndTime(req.GetEndTime())
	} else {
		update = update.ClearEndTime()
	}
	update = update.SetIsPickup(req.GetIsPickup()).SetIsDropoff(req.GetIsDropoff()).SetAgreementNoShopping(req.GetAgreementNoShopping()).SetAgreementAdjustable(req.GetAgreementAdjustable())
	// 新增字段
	if req.GetPoiId() != "" {
		update = update.SetPoiID(req.GetPoiId())
	} else {
		update = update.ClearPoiID()
	}
	if req.GetPoiName() != "" {
		update = update.SetPoiName(req.GetPoiName())
	} else {
		update = update.ClearPoiName()
	}
	update = update.SetIsEntering(req.GetIsEntering())
	if req.GetDurationMode() != "" {
		update = update.SetDurationMode(req.GetDurationMode())
	} else {
		update = update.ClearDurationMode()
	}
	if req.GetDurationHours() > 0 {
		update = update.SetDurationHours(int(req.GetDurationHours()))
	} else {
		update = update.ClearDurationHours()
	}
	if req.GetActivityFeatures() != "" {
		update = update.SetActivityFeatures(req.GetActivityFeatures())
	} else {
		update = update.ClearActivityFeatures()
	}
	if req.GetPickupLocation() != "" {
		update = update.SetPickupLocation(req.GetPickupLocation())
	} else {
		update = update.ClearPickupLocation()
	}
	if req.GetPickupAddress() != "" {
		update = update.SetPickupAddress(req.GetPickupAddress())
	} else {
		update = update.ClearPickupAddress()
	}
	if req.GetPickupLatitude() != 0 {
		update = update.SetPickupLatitude(req.GetPickupLatitude())
	} else {
		update = update.ClearPickupLatitude()
	}
	if req.GetPickupLongitude() != 0 {
		update = update.SetPickupLongitude(req.GetPickupLongitude())
	} else {
		update = update.ClearPickupLongitude()
	}
	if req.GetDropoffLocation() != "" {
		update = update.SetDropoffLocation(req.GetDropoffLocation())
	} else {
		update = update.ClearDropoffLocation()
	}
	if req.GetDropoffAddress() != "" {
		update = update.SetDropoffAddress(req.GetDropoffAddress())
	} else {
		update = update.ClearDropoffAddress()
	}
	if req.GetDropoffLatitude() != 0 {
		update = update.SetDropoffLatitude(req.GetDropoffLatitude())
	} else {
		update = update.ClearDropoffLatitude()
	}
	if req.GetDropoffLongitude() != 0 {
		update = update.SetDropoffLongitude(req.GetDropoffLongitude())
	} else {
		update = update.ClearDropoffLongitude()
	}
	update = update.SetTypeParams(req.GetTypeParams())
	stop, err = update.Save(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return itineraryStopMessage(stop), nil
}

func (s *ManagementService) DeleteItineraryStop(ctx context.Context, req *travel.DeleteItineraryStopRequest) (*travel.Empty, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetId() <= 0 || req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "id and productId are required")
	}
	_, err = s.client.Product.Query().Where(product.IDEQ(int(req.GetProductId())), product.TenantIDEQ(tenantID), product.MerchantIDEQ(merchantID)).Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "product not found")
		}
		return nil, err
	}
	err = s.client.ItineraryStop.DeleteOneID(int(req.GetId())).Where(itinerarystop.ProductIDEQ(req.GetProductId()), itinerarystop.TenantIDEQ(tenantID), itinerarystop.MerchantIDEQ(merchantID)).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, status.Error(codes.NotFound, "itinerary stop not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &travel.Empty{}, nil
}

func (s *ManagementService) ListItineraryStops(ctx context.Context, req *travel.ItineraryStopListRequest) (*travel.ItineraryStopListResponse, error) {
	tenantID, merchantID, err := scope(ctx)
	if err != nil {
		return nil, err
	}
	if req.GetProductId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "productId is required")
	}
	items, err := s.client.ItineraryStop.Query().Where(itinerarystop.ProductIDEQ(req.GetProductId()), itinerarystop.TenantIDEQ(tenantID), itinerarystop.MerchantIDEQ(merchantID)).Order(ent.Asc(itinerarystop.FieldSequence), ent.Asc(itinerarystop.FieldID)).All(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	out := &travel.ItineraryStopListResponse{}
	for _, item := range items {
		out.Items = append(out.Items, itineraryStopMessage(item))
	}
	return out, nil
}
