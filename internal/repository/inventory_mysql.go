package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/inventory"
	"gitee.com/meinongyihe/travel-rpc/ent/inventoryreservation"
)

type mysqlInventoryRepository struct { client *ent.Client }
func NewInventoryRepository(client *ent.Client) InventoryRepository { return &mysqlInventoryRepository{client: client} }

func (r *mysqlInventoryRepository) Check(ctx context.Context, tenantID, packageID int64, serviceDate, timeSlot string, quantity int) (*InventoryAvailability, error) {
	if tenantID <= 0 || packageID <= 0 || serviceDate == "" || quantity <= 0 { return nil, &ErrInvalidReservation{} }
	item, err := r.client.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.PackageIDEQ(packageID), inventory.ServiceDateEQ(serviceDate), inventory.TimeSlotEQ(timeSlot), inventory.StatusEQ("OPEN")).Only(ctx)
	if err != nil { return nil, err }
	return availability(item), nil
}

func (r *mysqlInventoryRepository) Reserve(ctx context.Context, tenantID, merchantID, packageID int64, serviceDate, timeSlot string, quantity int, reservationKey string) (*ReservationResult, error) {
	if tenantID <= 0 || merchantID <= 0 || packageID <= 0 || serviceDate == "" || quantity <= 0 || reservationKey == "" { return nil, &ErrInvalidReservation{} }
	var result *ReservationResult
	tx, err := r.client.Tx(ctx); if err != nil { return nil, err }
	defer func() { if result == nil { _ = tx.Rollback() } }()

	existing, err := tx.InventoryReservation.Query().Where(inventoryreservation.TenantIDEQ(tenantID), inventoryreservation.ReservationKeyEQ(reservationKey)).Only(ctx)
	if err == nil {
		if existing.MerchantID != merchantID || existing.Quantity != quantity || existing.Status == "RELEASED" || existing.Status == "EXPIRED" { _ = tx.Rollback(); return nil, &ErrReservationKeyConflict{} }
		inv, getErr := tx.Inventory.Get(ctx, int(existing.InventoryID)); if getErr != nil { _ = tx.Rollback(); return nil, getErr }
		if int64(inv.PackageID) != packageID || inv.ServiceDate != serviceDate || inv.TimeSlot != timeSlot { _ = tx.Rollback(); return nil, &ErrReservationKeyConflict{} }
		result = &ReservationResult{Reservation: existing, Remaining: inv.Capacity-inv.Reserved, UnitPrice: inv.UnitPrice, Currency: inv.Currency}
		if err = tx.Commit(); err != nil { return nil, err }; return result, nil
	}
	if !ent.IsNotFound(err) { _ = tx.Rollback(); return nil, err }

	item, err := tx.Inventory.Query().Where(inventory.TenantIDEQ(tenantID), inventory.MerchantIDEQ(merchantID), inventory.PackageIDEQ(packageID), inventory.ServiceDateEQ(serviceDate), inventory.TimeSlotEQ(timeSlot), inventory.StatusEQ("OPEN")).Only(ctx)
	if err != nil { _ = tx.Rollback(); return nil, err }
	if item.Capacity-item.Reserved < quantity { _ = tx.Rollback(); return nil, &ErrInsufficientInventory{Remaining:item.Capacity-item.Reserved} }
	updated, err := tx.Inventory.UpdateOneID(item.ID).Where(inventory.ReservedEQ(item.Reserved), inventory.StatusEQ("OPEN")).SetReserved(item.Reserved+quantity).Save(ctx)
	if err != nil { _ = tx.Rollback(); return nil, err }
	hold, err := tx.InventoryReservation.Create().SetTenantID(tenantID).SetMerchantID(merchantID).SetInventoryID(int64(updated.ID)).SetReservationKey(reservationKey).SetQuantity(quantity).SetStatus("RESERVED").SetExpiresAt(time.Now().Add(15*time.Minute)).Save(ctx)
	if err != nil { _ = tx.Rollback(); return nil, err }
	result = &ReservationResult{Reservation:hold, Remaining:updated.Capacity-updated.Reserved, UnitPrice:updated.UnitPrice, Currency:updated.Currency}
	if err = tx.Commit(); err != nil { return nil, err }; return result, nil
}

func (r *mysqlInventoryRepository) ConfirmReservation(ctx context.Context, tenantID, reservationID, orderID int64) error {
	if tenantID <= 0 || reservationID <= 0 || orderID <= 0 { return &ErrInvalidReservation{} }
	_, err := r.client.InventoryReservation.UpdateOneID(int(reservationID)).Where(inventoryreservation.TenantIDEQ(tenantID), inventoryreservation.StatusEQ("RESERVED"), inventoryreservation.ExpiresAtGT(time.Now())).SetStatus("CONFIRMED").SetOrderID(orderID).Save(ctx)
	return err
}

func (r *mysqlInventoryRepository) ReleaseReservation(ctx context.Context, tenantID, reservationID int64) error {
	tx, err := r.client.Tx(ctx); if err != nil { return err }
	hold, err := tx.InventoryReservation.Query().Where(inventoryreservation.IDEQ(int(reservationID)), inventoryreservation.TenantIDEQ(tenantID), inventoryreservation.StatusEQ("RESERVED")).Only(ctx)
	if err != nil { _ = tx.Rollback(); return err }
	item, err := tx.Inventory.Get(ctx, int(hold.InventoryID)); if err != nil { _ = tx.Rollback(); return err }
	if item.Reserved < hold.Quantity { _ = tx.Rollback(); return &ErrInventoryInvariant{} }
	if _, err = tx.Inventory.UpdateOneID(item.ID).SetReserved(item.Reserved-hold.Quantity).Save(ctx); err != nil { _ = tx.Rollback(); return err }
	if _, err = tx.InventoryReservation.UpdateOneID(hold.ID).SetStatus("RELEASED").Save(ctx); err != nil { _ = tx.Rollback(); return err }
	return tx.Commit()
}

func (r *mysqlInventoryRepository) ExpireReservations(ctx context.Context, nowUnix int64, limit int) (int, error) {
	if limit <= 0 || limit > 500 { limit = 500 }
	items, err := r.client.InventoryReservation.Query().Where(inventoryreservation.StatusEQ("RESERVED"), inventoryreservation.ExpiresAtLTE(time.Unix(nowUnix,0))).Limit(limit).All(ctx)
	if err != nil { return 0, err }
	count := 0
	for _, hold := range items { if err := r.ReleaseReservation(ctx, hold.TenantID, int64(hold.ID)); err == nil { count++ } }
	return count, nil
}

func availability(item *ent.Inventory) *InventoryAvailability { return &InventoryAvailability{InventoryID:int64(item.ID), Remaining:item.Capacity-item.Reserved, UnitPrice:item.UnitPrice, Currency:item.Currency} }

type ErrInsufficientInventory struct { Remaining int }
func (e *ErrInsufficientInventory) Error() string { return "insufficient inventory" }
type ErrInvalidReservation struct{}
func (e *ErrInvalidReservation) Error() string { return "invalid inventory reservation" }
type ErrReservationKeyConflict struct{}
func (e *ErrReservationKeyConflict) Error() string { return "reservation key conflicts with an existing reservation" }
type ErrInventoryInvariant struct{}
func (e *ErrInventoryInvariant) Error() string { return "inventory reservation invariant violated" }
