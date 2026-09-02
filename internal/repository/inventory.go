package repository

import (
	"context"

	"gitee.com/meinongyihe/travel-rpc/ent"
)

// InventoryRepository is the persistence boundary for availability and holds.
// Reserve must be implemented as one database transaction with an idempotency
// check on reservation_key; this interface deliberately keeps transaction
// semantics out of the RPC contract.
type InventoryRepository interface {
	Check(ctx context.Context, tenantID, packageID int64, serviceDate, timeSlot string, quantity int) (*InventoryAvailability, error)
	Reserve(ctx context.Context, tenantID, merchantID, packageID int64, serviceDate, timeSlot string, quantity int, reservationKey string) (*ReservationResult, error)
	ConfirmReservation(ctx context.Context, tenantID int64, reservationID int64, orderID int64) error
	ReleaseReservation(ctx context.Context, tenantID int64, reservationID int64) error
	ExpireReservations(ctx context.Context, nowUnix int64, limit int) (int, error)
}

type InventoryAvailability struct {
	InventoryID int64
	Remaining   int
	UnitPrice   int64
	Currency    string
}

type ReservationResult struct {
	Reservation *ent.InventoryReservation
	Remaining   int
	UnitPrice   int64
	Currency    string
}
