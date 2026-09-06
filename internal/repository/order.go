package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// OrderRepository is the persistence boundary for booking/order creation.
// Final amount and currency must be calculated from server-side product,
// package and inventory pricing; callers must not be treated as authoritative.
type OrderRepository interface {
	Create(ctx context.Context, input CreateOrderInput) (*ent.Order, error)
	GetByOrderNo(ctx context.Context, tenantID, merchantID int64, orderNo string) (*ent.Order, error)
	List(ctx context.Context, tenantID, merchantID int64, status string, page, pageSize int32) ([]*ent.Order, int64, error)
	CountByStatus(ctx context.Context, tenantID, merchantID int64, status string) (int64, error)
	ListByCustomer(ctx context.Context, tenantID, customerID int64, status string, page, pageSize int32) ([]*ent.Order, int64, error)
	ListTravelersByOrderID(ctx context.Context, orderID int64) ([]*ent.Traveler, error)
	ListItemsByOrderID(ctx context.Context, orderID int64) ([]*ent.OrderItem, error)
}

type CreateOrderInput struct {
	TenantID      int64
	MerchantID    int64
	ProductID     int64
	PackageID     int64
	CustomerID    *int64
	CustomerEmail string
	CustomerName  string
	CustomerPhone string
	Quantity      int
	ServiceDate   string
	TimeSlot      string
	UnitPrice     int64
	Currency      string
	Remark        string
	Travelers     []TravelerInput
}

// TravelerInput holds the data needed to persist an order traveler.
type TravelerInput struct {
	Name     string
	IdType   string
	IdNumber string
	Email    string
	Phone    string
}
