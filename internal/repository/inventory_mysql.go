package repository

import (
	"context"
	"time"

	"gitee.com/meinongyihe/travel-rpc/ent"
	"gitee.com/meinongyihe/travel-rpc/ent/inventory"
	"gitee.com/meinongyihe/travel-rpc/ent/inventoryreservation"
	"gitee.com/meinongyihe/travel-rpc/ent/predicate"
	"gitee.com/meinongyihe/travel-rpc/ent/enttest"
)

// mysqlInventoryRepository provides the transaction boundary for inventory.
type mysqlInventoryRepository struct {
	client *ent.Client
}

func NewInventoryRepository(client *ent.Client) InventoryRepository {
	return &mysqlInventoryRepository{client: client}
}

func (r *mysqlInventoryRepository) Check(ctx context.Context, tenantID, packageID int64, serviceDate, timeSlot string, quantity int) (*InventoryAvailability, error) {
	item, err := r.client.Inventory.Query().Where(
		inventory.TenantIDEQ(tenantID),
		inventory.PackageIDEQ(packageID),
		inventory.ServiceDateEQ(serviceDate),
		inventory.TimeSlotEQ(timeSlot),
		inventory.StatusEQ("OPEN"),
	).Only(ctx)
	if err != nil {
		return nil, err
	}
	return availability(item, quantity)
}

func (r *mysqlInventoryRepository) Reserve(ctx context.Context, tenantID, merchantID, packageID int64, serviceDate, timeSlot string, quantity int, reservationKey string) (*ReservationResult, error) {
	var result *ReservationResult
	err := ent.Tx(ctx, r.client, func(tx *ent.Tx) error {
		// Retry-safe: an existing reservation for the same tenant/key is returned as-is.
		existing, err := tx.InventoryReservation.Query().Where(
			inventoryreservation.TenantIDEQ(tenantID),
			inventoryreservation.ReservationKeyEQ(reservationKey),
		).Only(ctx)
		if err == nil {
			inv, getErr := tx.Inventory.Get(ctx, existing.InventoryID)
			if getErr != nil {
				return getErr
			}
			result = &ReservationResult{Reservation: existing, Remaining: inv.Capacity - inv.Reserved, UnitPrice: inv.UnitPrice, Currency: inv.Currency}
			return nil
		}
		if !ent.IsNotFound(err) {
			return err
		}

		item, err := tx.Inventory.Query().Where(
			inventory.TenantIDEQ(tenantID),
			inventory.PackageIDEQ(packageID),
			inventory.ServiceDateEQ(serviceDate),
			inventory.TimeSlotEQ(timeSlot),
			inventory.StatusEQ("OPEN"),
		).Only(ctx)
		if err != nil {
			return err
		}
		if item.Capacity-item.Reserved < quantity {
			return &ErrInsufficientInventory{Remaining: item.Capacity - item.Reserved}
		}
		updated, err := tx.Inventory.UpdateOneID(item.ID).Where(
			inventory.ReservedEQ(item.Reserved),
			inventory.StatusEQ("OPEN"),
		).SetReserved(item.Reserved + quantity).Save(ctx)
		if err != nil {
			return err
		}
		hold, err := tx.InventoryReservation.Create().
			SetTenantID(tenantID).
			SetMerchantID(merchantID).
			SetInventoryID(updated.ID).
			SetReservationKey(reservationKey).
			SetQuantity(quantity).
			SetStatus("RESERVED").
			SetExpiresAt(time.Now().Add(15 * time.Minute)).
			Save(ctx)
		if err != nil {
			return err
		}
		result = &ReservationResult{Reservation: hold, Remaining: updated.Capacity - updated.Reserved, UnitPrice: updated.UnitPrice, Currency: updated.Currency}
		return nil
	})
	return result, err
}

func (r *mysqlInventoryRepository) ConfirmReservation(ctx context.Context, tenantID int64, reservationID int64, orderID int64) error {
	_, err := r.client.InventoryReservation.UpdateOneID(reservationID).
		Where(inventoryreservation.TenantIDEQ(tenantID), inventoryreservation.StatusEQ("RESERVED")).
		SetStatus("CONFIRMED").SetOrderID(orderID).Save(ctx)
	return err
}

func (r *mysqlInventoryRepository) ReleaseReservation(ctx context.Context, tenantID int64, reservationID int64) error {
	return ent.Tx(ctx, r.client, func(tx *ent.Tx) error {
		hold, err := tx.InventoryReservation.Query().Where(
			inventoryreservation.IDEQ(reservationID),
			inventoryreservation.TenantIDEQ(tenantID),
			inventoryreservation.StatusEQ("RESERVED"),
		).Only(ctx)
		if err != nil {
			return err
		}
		item, err := tx.Inventory.Get(ctx, hold.InventoryID)
		if err != nil {
			return err
		}
		if item.Reserved < hold.Quantity {
			return nil
		}
		if _, err = tx.Inventory.UpdateOneID(item.ID).SetReserved(item.Reserved - hold.Quantity).Save(ctx); err != nil {
			return err
		}
		_, err = tx.InventoryReservation.UpdateOneID(hold.ID).SetStatus("RELEASED").Save(ctx)
		return err
	})
}

func (r *mysqlInventoryRepository) ExpireReservations(ctx context.Context, nowUnix int64, limit int) (int, error) {
	now := time.Unix(nowUnix, 0)
	if limit <= 0 || limit > 500 {
		limit = 500
	}
	items, err := r.client.InventoryReservation.Query().Where(
		inventoryreservation.StatusEQ("RESERVED"),
		inventoryreservation.ExpiresAtLTE(now),
	).Limit(limit).All(ctx)
	if err != nil {
		return 0, err
	}
	count := 0
	for _, hold := range items {
		if err := r.ReleaseReservation(ctx, hold.TenantID, hold.ID); err == nil {
			count++
		}
	}
	return count, nil
}

func availability(item *ent.Inventory, quantity int) (*InventoryAvailability, error) {
	remaining := item.Capacity - item.Reserved
	if quantity <= 0 {
		quantity = 1
	}
	return &InventoryAvailability{InventoryID: item.ID, Remaining: remaining, UnitPrice: item.UnitPrice, Currency: item.Currency}, nil
}

// ErrInsufficientInventory is returned when the requested quantity cannot be held.
type ErrInsufficientInventory struct{ Remaining int }
func (e *ErrInsufficientInventory) Error() string { return "insufficient inventory" }

var _ predicate.Inventory = inventory.And()
var _ = enttest.IsTest
