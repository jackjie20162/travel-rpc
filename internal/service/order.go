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
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

    var customerID *int64
    if id, e := auth.CustomerID(ctx); e == nil { customerID = &id }

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

    created, err := s.booking.CreateFromReservation(ctx, hold.Reservation.ID, repository.CreateOrderInput{
        TenantID: tenantID, MerchantID: merchantID, ProductID: req.GetProductId(), PackageID: req.GetPackageId(),
        CustomerID: customerID, CustomerEmail: req.GetCustomerEmail(), Quantity: int(req.GetQuantity()),
        ServiceDate: req.GetDate(), TimeSlot: req.GetTimeSlot(), UnitPrice: hold.UnitPrice, Currency: hold.Currency,
    })
    if err != nil {
        // Only release an unconfirmed hold. A retry that already completed the
        // booking must never release inventory belonging to a confirmed order.
        if hold.Reservation.Status == "RESERVED" { _ = s.inventory.ReleaseReservation(ctx, tenantID, hold.Reservation.ID) }
        switch err.(type) {
        case *repository.ErrReservationExpired:
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
    return toOrder(o), nil
}

func toOrder(o *ent.Order) *travel.Order {
    return &travel.Order{Id: o.ID, OrderNo: o.OrderNo, Status: o.Status, TotalAmount: o.TotalAmount, Currency: o.Currency}
}
