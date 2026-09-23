package service

import (
    "context"
    "strings"
    "time"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/internal/auth"
    "gitee.com/meinongyihe/travel-rpc/internal/imnotify"
    "gitee.com/meinongyihe/travel-rpc/internal/repository"
    "gitee.com/meinongyihe/travel-rpc/travel"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

type OrderService struct {
    travel.UnimplementedOrderServiceServer
    orders     repository.OrderRepository
    inventory  repository.InventoryRepository
    booking    repository.BookingRepository
    currencies repository.CurrencyRepository
}

func NewOrderService(orders repository.OrderRepository, inventory repository.InventoryRepository, booking repository.BookingRepository, currencies repository.CurrencyRepository) *OrderService {
    return &OrderService{orders: orders, inventory: inventory, booking: booking, currencies: currencies}
}

func (s *OrderService) Create(ctx context.Context, req *travel.CreateOrderRequest) (*travel.Order, error) {
    if req == nil || req.GetProductId() <= 0 || req.GetPackageId() <= 0 || req.GetDate() == "" || req.GetQuantity() <= 0 || req.GetReservationKey() == "" {
        return nil, status.Error(codes.InvalidArgument, "product, package, date, quantity and reservation key are required")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    // merchantID is optional for public orders
    merchantID, _ := auth.MerchantID(ctx)

    // user_id is the account owner who creates the order (账号归属人)
    var userID *int64
    if id, e := auth.AuthenticatedUserID(ctx); e == nil && id > 0 {
        userID = &id
    }

    // customer_id is the contact person (联系人), only set when X-Customer-ID header is present
    var customerID *int64
    if id, e := auth.CustomerID(ctx); e == nil && id != nil && *id > 0 {
        customerID = id
    }

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

    // 下单锁汇：结算真值为库存基准币金额(total)，按用户所选展示币种锁定汇率与展示金额。
    displayCurrency, rateMicro := s.resolveDisplayCurrency(ctx, hold.Currency, req.GetDisplayCurrency())
    displayAmountMinor := convertToMinor(hold.UnitPrice*int64(req.GetQuantity()), rateMicro)

    created, err := s.booking.CreateFromReservation(ctx, int64(hold.Reservation.ID), repository.CreateOrderInput{
        TenantID: tenantID, MerchantID: invMerchantID, ProductID: req.GetProductId(), PackageID: req.GetPackageId(),
        UserID: userID, CustomerID: customerID, CustomerEmail: req.GetCustomerEmail(), CustomerName: req.GetCustomerName(),
        CustomerPhone: req.GetCustomerPhone(), Quantity: int(req.GetQuantity()),
        ServiceDate: req.GetDate(), TimeSlot: req.GetTimeSlot(), UnitPrice: hold.UnitPrice, Currency: hold.Currency,
        Remark: req.GetRemark(), Travelers: toTravelerInputs(req.GetTravelers()),
        DisplayCurrency: displayCurrency, ExchangeRateMicro: rateMicro, DisplayAmountMinor: displayAmountMinor,
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
    // Use user_id (账号归属人) from metadata for filtering
    var userID int64
    if id, e := auth.AuthenticatedUserID(ctx); e == nil && id > 0 {
        userID = id
    } else if id, e := auth.CustomerID(ctx); e == nil && id != nil && *id > 0 {
        userID = *id
    } else {
        userID = req.GetCustomerId()
    }
    if userID <= 0 { return nil, status.Error(codes.InvalidArgument, "user_id is required") }
    page, pageSize := req.GetPage(), req.GetPageSize()
    if page <= 0 { page = 1 }
    if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
    items, total, err := s.orders.ListByUser(ctx, tenantID, userID, req.GetStatus(), page, pageSize)
    if err != nil { return nil, status.Error(codes.Internal, err.Error()) }
    resp := &travel.CustomerOrderListResponse{Total: total}
    for _, o := range items { resp.Items = append(resp.Items, toOrder(o)) }
    return resp, nil
}

func (s *OrderService) VerifyOrder(ctx context.Context, req *travel.VerifyOrderRequest) (*travel.Order, error) {
    if req == nil || req.GetOrderNo() == "" {
        return nil, status.Error(codes.InvalidArgument, "order number is required")
    }
    action := req.GetAction()
    if action != "CONFIRM" && action != "REJECT" {
        return nil, status.Error(codes.InvalidArgument, "action must be CONFIRM or REJECT")
    }
    if action == "REJECT" && req.GetReason() == "" {
        return nil, status.Error(codes.InvalidArgument, "reason is required when rejecting an order")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

    // Fetch the order first to validate current status
    o, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.NotFound, "order not found")
    }
    // Only PENDING_VERIFY orders can be verified/rejected (核销)
    if o.Status != "PENDING_VERIFY" {
        return nil, status.Error(codes.FailedPrecondition, "order is not in pending verify status")
    }

    var newStatus string
    var rejectReason string
    var verifiedAt int64
    if action == "CONFIRM" {
        newStatus = "COMPLETED"
        verifiedAt = time.Now().Unix()
    } else {
        newStatus = "REFUNDED"
        rejectReason = req.GetReason()
    }

    if err := s.orders.UpdateStatus(ctx, tenantID, merchantID, req.GetOrderNo(), newStatus, rejectReason, verifiedAt); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    // Re-fetch the updated order
    updated, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to fetch updated order")
    }
    // IM 推送核销结果（商户->客户）
    verifyMsg := "订单已核销，祝您旅途愉快"
    if action == "REJECT" {
        verifyMsg = "订单核销未通过，已退款。原因：" + req.GetReason()
    }
    imnotify.NotifyOrder(ctx, updated, true, verifyMsg)
    return toOrder(updated), nil
}

func (s *OrderService) AcceptOrder(ctx context.Context, req *travel.AcceptOrderRequest) (*travel.Order, error) {
    if req == nil || req.GetOrderNo() == "" {
        return nil, status.Error(codes.InvalidArgument, "order number is required")
    }
    action := req.GetAction()
    if action != "ACCEPT" && action != "REJECT" {
        return nil, status.Error(codes.InvalidArgument, "action must be ACCEPT or REJECT")
    }
    if action == "REJECT" && req.GetReason() == "" {
        return nil, status.Error(codes.InvalidArgument, "reason is required when rejecting an order")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

    // Fetch the order to validate current status
    o, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.NotFound, "order not found")
    }
    // Only PENDING_ACCEPTANCE orders can be accepted/rejected
    if o.Status != "PENDING_ACCEPTANCE" {
        return nil, status.Error(codes.FailedPrecondition, "order is not in pending acceptance status")
    }

    var newStatus string
    var rejectReason string
    if action == "ACCEPT" {
        newStatus = "PENDING_VERIFY"
    } else {
        newStatus = "CANCELLED"
        rejectReason = req.GetReason()
    }

    if err := s.orders.AcceptOrder(ctx, tenantID, merchantID, req.GetOrderNo(), newStatus, rejectReason); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    updated, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to fetch updated order")
    }
    // IM 推送接单结果（商户->客户）
    acceptMsg := "商家已接单，请等待出行"
    if action == "REJECT" {
        acceptMsg = "商家拒绝了订单。原因：" + req.GetReason()
    }
    imnotify.NotifyOrder(ctx, updated, true, acceptMsg)
    return toOrder(updated), nil
}

func (s *OrderService) RequestRefund(ctx context.Context, req *travel.RefundRequest) (*travel.Order, error) {
    if req == nil || req.GetOrderNo() == "" {
        return nil, status.Error(codes.InvalidArgument, "order number is required")
    }
    if req.GetReason() == "" {
        return nil, status.Error(codes.InvalidArgument, "reason is required")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    userID, err := auth.AuthenticatedUserID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

    // Fetch the order to validate current status
    o, err := s.orders.GetByOrderNo(ctx, tenantID, 0, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.NotFound, "order not found")
    }
    // Only PENDING_ACCEPTANCE or PENDING_VERIFY orders can request refund
    if o.Status != "PENDING_ACCEPTANCE" && o.Status != "PENDING_VERIFY" {
        return nil, status.Error(codes.FailedPrecondition, "order status does not allow refund request")
    }

    if err := s.orders.RequestRefund(ctx, tenantID, int64(userID), req.GetOrderNo(), req.GetReason()); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    // Re-fetch without merchantID since this is a customer operation
    updated, err := s.orders.GetByOrderNo(ctx, tenantID, o.MerchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to fetch updated order")
    }
    // IM 推送退款申请（客户->商户，提醒商户及时处理）
    imnotify.NotifyOrder(ctx, updated, false, "客户申请退款。原因："+req.GetReason())
    return toOrder(updated), nil
}

func (s *OrderService) HandleRefund(ctx context.Context, req *travel.HandleRefundRequest) (*travel.Order, error) {
    if req == nil || req.GetOrderNo() == "" {
        return nil, status.Error(codes.InvalidArgument, "order number is required")
    }
    action := req.GetAction()
    if action != "APPROVE" && action != "REJECT" {
        return nil, status.Error(codes.InvalidArgument, "action must be APPROVE or REJECT")
    }
    tenantID, err := auth.TenantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
    merchantID, err := auth.MerchantID(ctx)
    if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

    // Fetch the order to validate current status
    o, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.NotFound, "order not found")
    }
    // Only PENDING_REFUND orders can be handled
    if o.Status != "PENDING_REFUND" {
        return nil, status.Error(codes.FailedPrecondition, "order is not in pending refund status")
    }

    approved := action == "APPROVE"
    if err := s.orders.HandleRefund(ctx, tenantID, merchantID, req.GetOrderNo(), approved, req.GetReason()); err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }

    updated, err := s.orders.GetByOrderNo(ctx, tenantID, merchantID, req.GetOrderNo())
    if err != nil {
        return nil, status.Error(codes.Internal, "failed to fetch updated order")
    }
    // IM 推送退款审批结果（商户->客户）
    refundMsg := "您的退款申请已通过，款项将退回原支付账户"
    if !approved {
        refundMsg = "您的退款申请被驳回。原因：" + req.GetReason()
    }
    imnotify.NotifyOrder(ctx, updated, true, refundMsg)
    return toOrder(updated), nil
}

func (s *OrderService) CancelOrder(ctx context.Context, req *travel.OrderNoRequest) (*travel.Order, error) {
	if req == nil || req.GetOrderNo() == "" {
		return nil, status.Error(codes.InvalidArgument, "order number is required")
	}
	tenantID, err := auth.TenantID(ctx)
	if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }
	userID, err := auth.AuthenticatedUserID(ctx)
	if err != nil { return nil, status.Error(codes.Unauthenticated, err.Error()) }

	if err := s.orders.CancelOrder(ctx, tenantID, int64(userID), req.GetOrderNo()); err != nil {
		if strings.Contains(err.Error(), "cannot be cancelled") {
			return nil, status.Error(codes.FailedPrecondition, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	o, err := s.orders.GetByOrderNo(ctx, tenantID, 0, req.GetOrderNo())
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to fetch updated order")
	}
	return toOrder(o), nil
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
		RejectReason: o.RejectReason, VerifiedAt: o.VerifiedAt,
		DisplayCurrency: o.DisplayCurrency, ExchangeRateMicro: o.ExchangeRateMicro, DisplayAmountMinor: o.DisplayAmountMinor,
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

// resolveDisplayCurrency 解析下单展示币种并锁定汇率。当请求币种为空、等于基准币，
// 或查不到有效汇率时，回退到基准币并按平价(1e6)锁定，避免以错误汇率锁定金额。
func (s *OrderService) resolveDisplayCurrency(ctx context.Context, baseCurrency, requested string) (string, int64) {
    if baseCurrency == "" {
        baseCurrency = repository.BaseCurrencyCode
    }
    if requested == "" || requested == baseCurrency {
        return baseCurrency, repository.RateMicroScale
    }
    if s.currencies == nil {
        return baseCurrency, repository.RateMicroScale
    }
    rateMicro, found, err := s.currencies.LatestRateMicro(ctx, baseCurrency, requested)
    if err != nil || !found {
        return baseCurrency, repository.RateMicroScale
    }
    return requested, rateMicro
}

// convertToMinor 把基准币整数「元」金额按 rate_micro 换算为展示币种的最小货币单位(×100)。
// 采用整数四舍五入避免浮点误差：minor = round(total × rate_micro / 1e4)。
// （total 元 × rate = 展示币元；×100 = 展示币分；rate = rate_micro/1e6，故除以 1e4。）
func convertToMinor(totalBase, rateMicro int64) int64 {
    if totalBase <= 0 || rateMicro <= 0 {
        return 0
    }
    return (totalBase*rateMicro + 5000) / 10000
}
