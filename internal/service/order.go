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

type OrderService struct {
    travel.UnimplementedOrderServiceServer
    orders    repository.OrderRepository
    inventory repository.InventoryRepository
    booking   repository.BookingRepository
}

func NewOrderService(orders repository.OrderRepository, inventory repository.InventoryRepository, booking repository.BookingRepository) *OrderService {
    return &OrderService{orders: orders, inventory: inventory, booking: booking}
}

func (s *OrderService) Create(ctx context.Context, req *travel.CreateOrderRequest) (*travel.Order, error) {
    if req == nil || req.GetProductId() <= 0 || req.GetPackageId() <= 0 || req.GetDate() == "" || req.GetQuantity() <= 0 || req.GetReservationKey() == "" {
        return nil, status.Error(codes.InvalidArgument, "product, package, date, quantity and reservation key are required")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    // merchantID is optional for public orders
    merchantID, _ := auth.MerchantID(ctx)

    var customerID *int64
    if id, e := auth.CustomerID(ctx); e == nil { customerID = id }

    // Reservation is a short-lived inventory hold. The following booking
    // transaction atomically creates the order and confirms that hold.
    hold, err := s.inventory.Reserve(ctx, tenantID, merchantID, req.GetPackageId(), req.GetDate(), req.GetTimeSlot(), int(req.GetQuantity()), req.GetReservationKey())
    if err != nil {
        switch err.(type) {
        case *repository.ErrInsufficientInventory:
            return nil, status.Error(codes.ResourceExhausted, err.Error())
        case *repository.ErrReservationKeyConflict:
            return nil, status.Error(codes.AlreadyExists, err.Error())
        default:
            return nil, status.Error(codes.Internal, err.Error())
        }
    }

    // Use the inventory owner's merchantID for booking so that
    // CreateFromReservation can match the reservation and inventory records.
    invMerchantID := hold.InventoryMerchantID

    created, err := s.booking.CreateFromReservation(ctx, int64(hold.Reservation.ID), repository.CreateOrderInput{
        TenantID: tenantID, MerchantID: invMerchantID, ProductID: req.GetProductId(), PackageID: req.GetPackageId(),
        CustomerID: customerID, CustomerEmail: req.GetCustomerEmail(), CustomerName: req.GetCustomerName(),
        CustomerPhone: req.GetCustomerPhone(), Quantity: int(req.GetQuantity()),
        ServiceDate: req.GetDate(), TimeSlot: req.GetTimeSlot(), UnitPrice: hold.UnitPrice, Currency: hold.Currency,
        Remark: req.GetRemark(), Travelers: toTravelerInputs(req.GetTravelers()),
    })
    if err != nil {
        // Only release an unconfirmed hold. A retry that already completed the
        // booking must never release inventory belonging to a confirmed order.
        if hold.Reservation.Status == "RESERVED" { _ = s.inventory.ReleaseReservation(ctx, tenantID, int64(hold.Reservation.ID)) }
        switch err.(type) {
        case *repository.ErrReservationExpired:
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        case *repository.ErrProductNotPublished:
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        case *repository.ErrProductInvariant:
            return nil, status.Error(codes.FailedPrecondition, err.Error())
        case *repository.ErrReservationNotActive, *repository.ErrReservationKeyConflict:
            return nil, status.Error(codes.Aborted, err.Error())
        default:
            return nil, status.Error(codes.Internal, err.Error())
        }
    }
    return toOrder(created), nil
}

func (s *OrderService) Get(ctx context.Context, req *travel.OrderNoRequest) (*travel.Order, error) {
    if req == nil || req.GetOrderNo() == "" { return nil, status.Error(codes.InvalidArgument, "order number is required") }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    o, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil { return nil, status.Error(codes.NotFound, "order not found") }
	travelers, _ := s.orders.ListTravelersByOrderID(ctx, int64(o.ID))
	items, _ := s.orders.ListItemsByOrderID(ctx, int64(o.ID))
	return toOrderWithTravelersAndItems(o, travelers, items), nil
}

func (s *OrderService) ListMerchantOrders(ctx context.Context, req *travel.MerchantOrderListRequest) (*travel.MerchantOrderListResponse, error) {
    if req == nil { return nil, status.Error(codes.InvalidArgument, "request is nil") }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    page, pageSize := req.GetPage(), req.GetPageSize()
    if page <= 0 { page = 1 }
    if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
    items, total, err := s.orders.List(ctx, tenantID, merchantID, req.GetStatus(), page, pageSize)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    resp := &travel.MerchantOrderListResponse{Total: total}
    for _, o := range items { resp.Items = append(resp.Items, toOrder(o)) }
    return resp, nil
}

func (s *OrderService) ListCustomerOrders(ctx context.Context, req *travel.CustomerOrderListRequest) (*travel.CustomerOrderListResponse, error) {
    if req == nil { return nil, status.Error(codes.InvalidArgument, "request is nil") }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    customerID := req.GetCustomerId()
    if customerID <= 0 { return nil, status.Error(codes.InvalidArgument, "customer_id is required") }
    page, pageSize := req.GetPage(), req.GetPageSize()
    if page <= 0 { page = 1 }
    if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
    items, total, err := s.orders.ListByCustomer(ctx, tenantID, customerID, req.GetStatus(), page, pageSize)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    resp := &travel.CustomerOrderListResponse{Total: total}
    for _, o := range items { resp.Items = append(resp.Items, toOrder(o)) }
    return resp, nil
}

func toOrder(o *ent.Order) *travel.Order {
	return &travel.Order{
        Id: int64(o.ID), OrderNo: o.OrderNo, Status: o.Status,
        TotalAmount: o.TotalAmount, Currency: o.Currency,
        CustomerEmail: o.CustomerEmail, PaymentStatus: o.PaymentStatus,
        CustomerId: o.CustomerID, CustomerName: o.CustomerName,
		CustomerPhone: o.CustomerPhone, Remark: o.Remark,
		ProductName: o.ProductName, PackageName: o.PackageName,
		ServiceDate: o.ServiceDate, TimeSlot: o.TimeSlot,
	}
}

func toOrderWithTravelers(o *ent.Order, travelers []*ent.Traveler) *travel.Order {
    result := toOrder(o)
	for _, t := range travelers {
		result.Travelers = append(result.Travelers, &travel.Traveler{
			Id: int64(t.ID), Name: t.Name, IdType: t.IDType, IdNumber: t.IDNumber, Phone: t.Phone,
		})
	}
	return result
}

func toOrderWithTravelersAndItems(o *ent.Order, travelers []*ent.Traveler, items []*ent.OrderItem) *travel.Order {
	result := toOrderWithTravelers(o, travelers)
	for _, item := range items {
		result.Items = append(result.Items, &travel.OrderItem{
			Id: int64(item.ID), ProductId: item.ProductID, PackageId: item.PackageID,
			TravelerId: item.TravelerID, Quantity: int32(item.Quantity), UnitPrice: item.UnitPrice,
			TotalAmount: item.TotalAmount, ServiceDate: item.ServiceDate, TimeSlot: item.TimeSlot,
			ProductCode: item.ProductCode, ProductName: item.ProductName,
			PackageCode: item.PackageCode, PackageName: item.PackageName,
		})
	}
	return result
}

func toTravelerInputs(ts []*travel.Traveler) []repository.TravelerInput {
    if len(ts) == 0 { return nil }
    out := make([]repository.TravelerInput, len(ts))
    for i, t := range ts {
        out[i] = repository.TravelerInput{
            Name: t.GetName(), IdType: t.GetIdType(), IdNumber: t.GetIdNumber(), Phone: t.GetPhone(),
        }
    }
    return out
}
