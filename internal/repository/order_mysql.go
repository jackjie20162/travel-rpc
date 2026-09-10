package repository

import (
	"context"
	"fmt"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/order"
	"gitee.com/meinongyihe/travel-rpc/ent/orderitem"
	"gitee.com/meinongyihe/travel-rpc/ent/predicate"
	"gitee.com/meinongyihe/travel-rpc/ent/traveler"
)

type mysqlOrderRepository struct{ client *ent.Client }

func NewOrderRepository(client *ent.Client) OrderRepository {
	return &mysqlOrderRepository{client: client}
}

func (r *mysqlOrderRepository) Create(ctx context.Context, input CreateOrderInput) (*ent.Order, error) {
	if input.TenantID <= 0 || input.MerchantID <= 0 || input.Quantity <= 0 || input.UnitPrice < 0 || input.Currency == "" || input.ProductID <= 0 || input.PackageID <= 0 || input.ServiceDate == "" {
		return nil, &ErrInvalidOrder{}
	}
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, err
	}
	total := input.UnitPrice * int64(input.Quantity)
	builder := tx.Order.Create().SetTenantID(input.TenantID).SetMerchantID(input.MerchantID).
		SetOrderNo(newOrderNo()).SetTotalAmount(total).SetCurrency(input.Currency).
		SetStatus("PENDING_PAYMENT").SetPaymentStatus("PENDING")
	if input.UserID != nil {
		builder.SetUserID(*input.UserID)
	}
	if input.CustomerID != nil {
		builder.SetCustomerID(*input.CustomerID)
	}
	if input.CustomerEmail != "" {
		builder.SetCustomerEmail(input.CustomerEmail)
	}
	// 下单锁汇字段：结算真值仍为 total(基准币)，以下仅用于用户侧展示。
	if input.DisplayCurrency != "" {
		builder.SetDisplayCurrency(input.DisplayCurrency)
	}
	if input.ExchangeRateMicro > 0 {
		builder.SetExchangeRateMicro(input.ExchangeRateMicro)
	}
	if input.DisplayAmountMinor > 0 {
		builder.SetDisplayAmountMinor(input.DisplayAmountMinor)
	}
	item, err := builder.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if _, err = tx.OrderItem.Create().SetOrderID(int64(item.ID)).SetProductID(input.ProductID).SetPackageID(input.PackageID).
		SetQuantity(input.Quantity).SetUnitPrice(input.UnitPrice).SetTotalAmount(total).
		SetServiceDate(input.ServiceDate).SetTimeSlot(input.TimeSlot).Save(ctx); err != nil {
		_ = tx.Rollback()
		return nil, err
	}
	if err = tx.Commit(); err != nil {
		return nil, err
	}
	return item, nil
}

func (r *mysqlOrderRepository) GetByOrderNo(ctx context.Context, tenantID, merchantID int64, orderNo string) (*ent.Order, error) {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.OrderNoEQ(orderNo)}
	if merchantID > 0 {
		preds = append(preds, order.MerchantIDEQ(merchantID))
	}
	return r.client.Order.Query().Where(preds...).Only(ctx)
}

func (r *mysqlOrderRepository) List(ctx context.Context, tenantID, merchantID int64, status string, page, pageSize int32) ([]*ent.Order, int64, error) {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.MerchantIDEQ(merchantID)}
	if status != "" {
		if status == "PAID" {
			preds = append(preds, order.PaymentStatusEQ(status))
		} else {
			preds = append(preds, order.StatusEQ(status))
		}
	}
	q := r.client.Order.Query().Where(preds...)
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page > 0 && pageSize > 0 {
		q = q.Offset(int((page - 1) * pageSize)).Limit(int(pageSize))
	}
	items, err := q.Order(ent.Desc(order.FieldID)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int64(total), nil
}

func (r *mysqlOrderRepository) CountByStatus(ctx context.Context, tenantID, merchantID int64, status string) (int64, error) {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.MerchantIDEQ(merchantID)}
	if status != "" {
		if status == "PAID" {
			preds = append(preds, order.PaymentStatusEQ(status))
		} else {
			preds = append(preds, order.StatusEQ(status))
		}
	}
	n, err := r.client.Order.Query().Where(preds...).Count(ctx)
	return int64(n), err
}

func (r *mysqlOrderRepository) ListByUser(ctx context.Context, tenantID, userID int64, status string, page, pageSize int32) ([]*ent.Order, int64, error) {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.UserIDEQ(userID)}
	if status != "" {
		if status == "PAID" {
			preds = append(preds, order.PaymentStatusEQ(status))
		} else {
			preds = append(preds, order.StatusEQ(status))
		}
	}
	q := r.client.Order.Query().Where(preds...)
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page > 0 && pageSize > 0 {
		q = q.Offset(int((page - 1) * pageSize)).Limit(int(pageSize))
	}
	items, err := q.Order(ent.Desc(order.FieldID)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int64(total), nil
}

func (r *mysqlOrderRepository) ListByCustomer(ctx context.Context, tenantID, customerID int64, status string, page, pageSize int32) ([]*ent.Order, int64, error) {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.CustomerIDEQ(customerID)}
	if status != "" {
		if status == "PAID" {
			preds = append(preds, order.PaymentStatusEQ(status))
		} else {
			preds = append(preds, order.StatusEQ(status))
		}
	}
	q := r.client.Order.Query().Where(preds...)
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, 0, err
	}
	if page > 0 && pageSize > 0 {
		q = q.Offset(int((page - 1) * pageSize)).Limit(int(pageSize))
	}
	items, err := q.Order(ent.Desc(order.FieldID)).All(ctx)
	if err != nil {
		return nil, 0, err
	}
	return items, int64(total), nil
}

func (r *mysqlOrderRepository) ListTravelersByOrderID(ctx context.Context, orderID int64) ([]*ent.Traveler, error) {
	return r.client.Traveler.Query().Where(traveler.OrderIDEQ(orderID)).All(ctx)
}

func (r *mysqlOrderRepository) ListItemsByOrderID(ctx context.Context, orderID int64) ([]*ent.OrderItem, error) {
	return r.client.OrderItem.Query().Where(orderitem.OrderIDEQ(orderID)).Order(ent.Asc(orderitem.FieldID)).All(ctx)
}

func (r *mysqlOrderRepository) UpdateStatus(ctx context.Context, tenantID, merchantID int64, orderNo string, newStatus string, rejectReason string, verifiedAt int64) error {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.MerchantIDEQ(merchantID), order.OrderNoEQ(orderNo)}
	o, err := r.client.Order.Query().Where(preds...).Only(ctx)
	if err != nil {
		return err
	}
	u := r.client.Order.UpdateOneID(o.ID).SetStatus(newStatus)
	if rejectReason != "" {
		u = u.SetRejectReason(rejectReason)
	}
	if verifiedAt > 0 {
		u = u.SetVerifiedAt(verifiedAt)
	}
	_, err = u.Save(ctx)
	return err
}

func (r *mysqlOrderRepository) AcceptOrder(ctx context.Context, tenantID, merchantID int64, orderNo string, newStatus string, rejectReason string) error {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.MerchantIDEQ(merchantID), order.OrderNoEQ(orderNo)}
	o, err := r.client.Order.Query().Where(preds...).Only(ctx)
	if err != nil {
		return err
	}
	u := r.client.Order.UpdateOneID(o.ID).SetStatus(newStatus)
	if rejectReason != "" {
		u = u.SetRejectReason(rejectReason)
	}
	_, err = u.Save(ctx)
	return err
}

func (r *mysqlOrderRepository) RequestRefund(ctx context.Context, tenantID int64, userID int64, orderNo string, reason string) error {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.UserIDEQ(userID), order.OrderNoEQ(orderNo)}
	o, err := r.client.Order.Query().Where(preds...).Only(ctx)
	if err != nil {
		return err
	}
	// Save current status as prev_status before transitioning to PENDING_REFUND
	_, err = r.client.Order.UpdateOneID(o.ID).
		SetPrevStatus(o.Status).
		SetStatus("PENDING_REFUND").
		SetRejectReason(reason).
		Save(ctx)
	return err
}

func (r *mysqlOrderRepository) HandleRefund(ctx context.Context, tenantID, merchantID int64, orderNo string, approved bool, reason string) error {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.MerchantIDEQ(merchantID), order.OrderNoEQ(orderNo)}
	o, err := r.client.Order.Query().Where(preds...).Only(ctx)
	if err != nil {
		return err
	}
	if approved {
		_, err = r.client.Order.UpdateOneID(o.ID).
			SetStatus("REFUNDED").
			SetPaymentStatus("REFUNDED").
			SetRejectReason(reason).
			Save(ctx)
	} else {
		// Restore previous status
		prevStatus := o.PrevStatus
		if prevStatus == "" {
			prevStatus = "PENDING_VERIFY"
		}
		_, err = r.client.Order.UpdateOneID(o.ID).
			SetStatus(prevStatus).
			ClearPrevStatus().
			SetRejectReason(reason).
			Save(ctx)
	}
	return err
}

func (r *mysqlOrderRepository) CancelOrder(ctx context.Context, tenantID int64, userID int64, orderNo string) error {
	preds := []predicate.Order{order.TenantIDEQ(tenantID), order.UserIDEQ(userID), order.OrderNoEQ(orderNo)}
	o, err := r.client.Order.Query().Where(preds...).Only(ctx)
	if err != nil {
		return err
	}
	if o.Status != "PENDING_PAYMENT" {
		return fmt.Errorf("order cannot be cancelled in %s status", o.Status)
	}
	_, err = r.client.Order.UpdateOneID(o.ID).SetStatus("CANCELLED").Save(ctx)
	return err
}

func newOrderNo() string {
	return fmt.Sprintf("TRV%s", time.Now().UTC().Format("20060102150405.000000000"))
}

type ErrInvalidOrder struct{}

func (e *ErrInvalidOrder) Error() string { return "invalid order" }
