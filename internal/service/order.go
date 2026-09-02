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

type OrderService struct { travel.UnimplementedOrderServiceServer; orders repository.OrderRepository; inventory repository.InventoryRepository }
func NewOrderService(orders repository.OrderRepository, inventory repository.InventoryRepository) *OrderService { return &OrderService{orders:orders,inventory:inventory} }
func (s *OrderService) Create(ctx context.Context, req *travel.CreateOrderRequest) (*travel.Order, error) {
	if req == nil || req.GetProductId()<=0 || req.GetPackageId()<=0 || req.GetDate()=="" || req.GetQuantity()<=0 || req.GetReservationKey()=="" { return nil,status.Error(codes.InvalidArgument,"product, package, date, quantity and reservation key are required") }
	tenantID,err:=auth.TenantID(ctx); if err!=nil{return nil,status.Error(codes.Unauthenticated,err.Error())}; merchantID,err:=auth.MerchantID(ctx); if err!=nil{return nil,status.Error(codes.Unauthenticated,err.Error())}
	var customerID *int64; if id,e:=auth.CustomerID(ctx); e==nil { customerID=&id }
	hold,err:=s.inventory.Reserve(ctx,tenantID,merchantID,req.GetPackageId(),req.GetDate(),req.GetTimeSlot(),int(req.GetQuantity()),req.GetReservationKey()); if err!=nil{return nil,status.Error(codes.ResourceExhausted,err.Error())}
	created,err:=s.orders.Create(ctx,repository.CreateOrderInput{TenantID:tenantID,MerchantID:merchantID,ProductID:req.GetProductId(),PackageID:req.GetPackageId(),CustomerID:customerID,CustomerEmail:req.GetCustomerEmail(),Quantity:int(req.GetQuantity()),ServiceDate:req.GetDate(),TimeSlot:req.GetTimeSlot(),UnitPrice:hold.UnitPrice,Currency:hold.Currency})
	if err!=nil{_ = s.inventory.ReleaseReservation(ctx,tenantID,hold.Reservation.ID); return nil,status.Error(codes.Internal,err.Error())}
	if err=s.inventory.ConfirmReservation(ctx,tenantID,hold.Reservation.ID,created.ID);err!=nil{return nil,status.Error(codes.Internal,err.Error())}
	return toOrder(created),nil
}
func (s *OrderService) Get(ctx context.Context, req *travel.OrderNoRequest) (*travel.Order,error) {
	if req==nil || req.GetOrderNo()=="" { return nil,status.Error(codes.InvalidArgument,"order number is required") }
	tenantID,err:=auth.TenantID(ctx); if err!=nil{return nil,status.Error(codes.Unauthenticated,err.Error())}; o,err:=s.orders.GetByOrderNo(ctx,tenantID,req.GetOrderNo()); if err!=nil{return nil,status.Error(codes.NotFound,"order not found")}; return toOrder(o),nil
}
func toOrder(o *ent.Order) *travel.Order { return &travel.Order{Id:o.ID,OrderNo:o.OrderNo,Status:o.Status,TotalAmount:o.TotalAmount,Currency:o.Currency} }
