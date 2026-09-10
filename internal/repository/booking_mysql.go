package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/inventoryreservation"
	"gitee.com/meinongyihe/travel-rpc/ent/product"
	"gitee.com/meinongyihe/travel-rpc/ent/productpackage"
)

type mysqlBookingRepository struct { client *ent.Client }
func NewBookingRepository(client *ent.Client) BookingRepository { return &mysqlBookingRepository{client: client} }

func (r *mysqlBookingRepository) CreateFromReservation(ctx context.Context, reservationID int64, input CreateOrderInput) (*ent.Order, error) {
	if reservationID <= 0 || input.TenantID <= 0 || input.MerchantID <= 0 || input.ProductID <= 0 || input.PackageID <= 0 || input.Quantity <= 0 || input.ServiceDate == "" { return nil, &ErrInvalidOrder{} }
	tx, err := r.client.Tx(ctx); if err != nil { return nil, err }
	committed := false
	defer func() { if !committed { _ = tx.Rollback() } }()
	hold, err := tx.InventoryReservation.Query().Where(inventoryreservation.IDEQ(int(reservationID)), inventoryreservation.TenantIDEQ(input.TenantID), inventoryreservation.MerchantIDEQ(input.MerchantID)).Only(ctx)
	if err != nil { return nil, err }
	if hold.Status == "CONFIRMED" {
		if hold.OrderID <= 0 { return nil, &ErrReservationNotActive{} }
		order, err := tx.Order.Get(ctx, int(hold.OrderID)); if err != nil { return nil, err }
		if order.TenantID != input.TenantID || order.MerchantID != input.MerchantID { return nil, &ErrReservationNotActive{} }
		if err = tx.Commit(); err != nil { return nil, err }; committed = true; return order, nil
	}
	if hold.Status != "RESERVED" { return nil, &ErrReservationNotActive{} }
	if !hold.ExpiresAt.After(time.Now()) { return nil, &ErrReservationExpired{} }
	if hold.Quantity != input.Quantity { return nil, &ErrReservationKeyConflict{} }
	inv, err := tx.Inventory.Get(ctx, int(hold.InventoryID)); if err != nil { return nil, err }
	if inv.TenantID != input.TenantID || inv.MerchantID != input.MerchantID || int64(inv.PackageID) != input.PackageID || inv.ServiceDate != input.ServiceDate || inv.TimeSlot != input.TimeSlot || inv.Status != "OPEN" { return nil, &ErrReservationKeyConflict{} }
	prod, err := tx.Product.Get(ctx, int(input.ProductID)); if err != nil { return nil, err }
	if prod.TenantID != input.TenantID || prod.MerchantID != input.MerchantID {
		return nil, &ErrProductInvariant{Msg: "product does not belong to the same tenant/merchant"}
	}
	if prod.Status != "PUBLISHED" {
		return nil, &ErrProductNotPublished{Status: prod.Status}
	}
	if prod.Currency != inv.Currency {
		return nil, &ErrProductInvariant{Msg: "product currency does not match inventory currency"}
	}
	pkg, err := tx.ProductPackage.Query().Where(productpackage.IDEQ(int(input.PackageID)), productpackage.ProductIDEQ(input.ProductID), productpackage.TenantIDEQ(input.TenantID), productpackage.MerchantIDEQ(input.MerchantID)).Only(ctx)
	if err != nil { return nil, &ErrProductInvariant{Msg: "package does not belong to the same product or merchant"} }
	total := inv.UnitPrice * int64(input.Quantity)
	builder := tx.Order.Create().SetTenantID(input.TenantID).SetMerchantID(input.MerchantID).SetOrderNo(newOrderNo()).SetTotalAmount(total).SetCurrency(inv.Currency).SetStatus("PENDING_PAYMENT").SetPaymentStatus("PENDING")
	if input.UserID != nil { builder.SetUserID(*input.UserID) }
	if input.CustomerID != nil { builder.SetCustomerID(*input.CustomerID) }
	if input.CustomerEmail != "" { builder.SetCustomerEmail(input.CustomerEmail) }
	if input.CustomerName != "" { builder.SetCustomerName(input.CustomerName) }
	if input.CustomerPhone != "" { builder.SetCustomerPhone(input.CustomerPhone) }
	if input.Remark != "" { builder.SetRemark(input.Remark) }
	// 下单锁汇字段：结算真值仍为 total(基准币)，以下仅用于用户侧展示。
	if input.DisplayCurrency != "" { builder.SetDisplayCurrency(input.DisplayCurrency) }
	if input.ExchangeRateMicro > 0 { builder.SetExchangeRateMicro(input.ExchangeRateMicro) }
	if input.DisplayAmountMinor > 0 { builder.SetDisplayAmountMinor(input.DisplayAmountMinor) }
	builder.SetProductName(prod.Title).SetPackageName(pkg.Name).SetServiceDate(input.ServiceDate).SetTimeSlot(input.TimeSlot)
	order, err := builder.Save(ctx); if err != nil { return nil, err }
	// Create traveler records first so each order item can point to the exact traveler.
	travelerIDs := make([]int64, 0, len(input.Travelers))
	for _, t := range input.Travelers {
		tb := tx.Traveler.Create().SetTenantID(input.TenantID).SetOrderID(int64(order.ID)).SetName(t.Name)
		if t.IdType != "" { tb.SetIDType(t.IdType) }
		if t.IdNumber != "" { tb.SetIDNumber(t.IdNumber) }
		if t.Email != "" { tb.SetEmail(t.Email) }
		if t.Phone != "" { tb.SetPhone(t.Phone) }
		traveler, saveErr := tb.Save(ctx); if saveErr != nil { return nil, saveErr }
		travelerIDs = append(travelerIDs, int64(traveler.ID))
	}
	itemBuilder := func() *ent.OrderItemCreate {
		return tx.OrderItem.Create().SetOrderID(int64(order.ID)).SetProductID(input.ProductID).SetPackageID(input.PackageID).SetQuantity(1).SetUnitPrice(inv.UnitPrice).SetTotalAmount(inv.UnitPrice).SetServiceDate(input.ServiceDate).SetTimeSlot(input.TimeSlot).SetProductCode(prod.Code).SetProductName(prod.Title).SetPackageCode(pkg.Code).SetPackageName(pkg.Name)
	}
	if len(travelerIDs) == 0 {
		if _, err = itemBuilder().SetQuantity(input.Quantity).SetTotalAmount(total).Save(ctx); err != nil { return nil, err }
	} else {
		for _, travelerID := range travelerIDs {
			if _, err = itemBuilder().SetTravelerID(travelerID).Save(ctx); err != nil { return nil, err }
		}
	}
	if _, err = tx.InventoryReservation.UpdateOneID(hold.ID).Where(inventoryreservation.StatusEQ("RESERVED"), inventoryreservation.ExpiresAtGT(time.Now())).SetStatus("CONFIRMED").SetOrderID(int64(order.ID)).Save(ctx); err != nil { return nil, err }
	if err = tx.Commit(); err != nil { return nil, err }; committed = true
	return order, nil
}

var _ = product.StatusEQ
