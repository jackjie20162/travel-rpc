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
	GetByOrderNo(ctx context.Context, tenantID int64, orderNo string) (*ent.Order, error)
}

type CreateOrderInput struct {
	TenantID      int64
	MerchantID    int64
	ProductID     int64
	PackageID     int64
	CustomerID    *int64
	CustomerEmail string
	Quantity      int
	ServiceDate   string
	TimeSlot      string
	UnitPrice     int64
	Currency      string
}
