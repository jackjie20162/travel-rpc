package repository

import (
	"context"
	"fmt"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/order"
)

type mysqlOrderRepository struct { client *ent.Client }

func NewOrderRepository(client *ent.Client) OrderRepository { return &mysqlOrderRepository{client:client} }

func (r *mysqlOrderRepository) Create(ctx context.Context, input CreateOrderInput) (*ent.Order, error) {
	if input.TenantID <= 0 || input.MerchantID <= 0 || input.Quantity <= 0 || input.UnitPrice < 0 || input.Currency == "" || input.ProductID <= 0 || input.PackageID <= 0 || input.ServiceDate == "" { return nil, &ErrInvalidOrder{} }
	tx, err := r.client.Tx(ctx)
	if err != nil { return nil, err }
	total := input.UnitPrice * int64(input.Quantity)
	builder := tx.Order.Create().SetTenantID(input.TenantID).SetMerchantID(input.MerchantID).
		SetOrderNo(newOrderNo()).SetTotalAmount(total).SetCurrency(input.Currency).
		SetStatus("PENDING_PAYMENT").SetPaymentStatus("PENDING")
	if input.CustomerID != nil { builder.SetCustomerID(*input.CustomerID) }
	if input.CustomerEmail != "" { builder.SetCustomerEmail(input.CustomerEmail) }
	item, err := builder.Save(ctx)
	if err != nil { _ = tx.Rollback(); return nil, err }
	if _, err = tx.OrderItem.Create().SetOrderID(item.ID).SetProductID(int(input.ProductID)).SetPackageID(int(input.PackageID)).
		SetQuantity(input.Quantity).SetUnitPrice(input.UnitPrice).SetTotalAmount(total).
		SetServiceDate(input.ServiceDate).SetTimeSlot(input.TimeSlot).Save(ctx); err != nil {
		_ = tx.Rollback(); return nil, err
	}
	if err = tx.Commit(); err != nil { return nil, err }
	return item, nil
}

func (r *mysqlOrderRepository) GetByOrderNo(ctx context.Context, tenantID, merchantID int64, orderNo string) (*ent.Order, error) {
	return r.client.Order.Query().Where(
		order.TenantIDEQ(tenantID),
		order.MerchantIDEQ(merchantID),
		order.OrderNoEQ(orderNo),
	).Only(ctx)
}

func newOrderNo() string { return fmt.Sprintf("TRV%s", time.Now().UTC().Format("20060102150405.000000000")) }

type ErrInvalidOrder struct{}
func (e *ErrInvalidOrder) Error() string { return "invalid order" }
