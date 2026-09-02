package repository

import (
    "context"
    "time"

    "gitee.com/meinongyihe/travel-rpc/ent"
    "gitee.com/meinongyihe/travel-rpc/ent/inventoryreservation"
    "gitee.com/meinongyihe/travel-rpc/ent/product"
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
    if prod.TenantID != input.TenantID || prod.MerchantID != input.MerchantID || prod.Status != "ACTIVE" || prod.Currency != inv.Currency { return nil, &ErrReservationKeyConflict{} }
    total := inv.UnitPrice * int64(input.Quantity)
    builder := tx.Order.Create().SetTenantID(input.TenantID).SetMerchantID(input.MerchantID).SetOrderNo(newOrderNo()).SetTotalAmount(total).SetCurrency(inv.Currency).SetStatus("PENDING_PAYMENT").SetPaymentStatus("PENDING")
    if input.CustomerID != nil { builder.SetCustomerID(*input.CustomerID) }
    if input.CustomerEmail != "" { builder.SetCustomerEmail(input.CustomerEmail) }
    order, err := builder.Save(ctx); if err != nil { return nil, err }
    if _, err = tx.OrderItem.Create().SetOrderID(int64(order.ID)).SetProductID(input.ProductID).SetPackageID(input.PackageID).SetQuantity(input.Quantity).SetUnitPrice(inv.UnitPrice).SetTotalAmount(total).SetServiceDate(input.ServiceDate).SetTimeSlot(input.TimeSlot).Save(ctx); err != nil { return nil, err }
    if _, err = tx.InventoryReservation.UpdateOneID(int(hold.ID)).Where(inventoryreservation.StatusEQ("RESERVED"), inventoryreservation.ExpiresAtGT(time.Now())).SetStatus("CONFIRMED").SetOrderID(int64(order.ID)).Save(ctx); err != nil { return nil, err }
    if err = tx.Commit(); err != nil { return nil, err }; committed = true
    return order, nil
}

var _ = product.StatusEQ
