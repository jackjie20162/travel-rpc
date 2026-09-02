package repository

import (
    "context"

    "gitee.com/meinongyihe/travel-rpc/ent"
)

// BookingRepository owns the atomic transition from an inventory reservation
// to a durable order. The operation is intentionally inside travel-rpc so the
// Travel service remains self-contained and retry-safe.
type BookingRepository interface {
    CreateFromReservation(ctx context.Context, reservationID int64, input CreateOrderInput) (*ent.Order, error)
}

type ErrReservationNotActive struct{}
func (e *ErrReservationNotActive) Error() string { return "inventory reservation is not active" }

type ErrReservationExpired struct{}
func (e *ErrReservationExpired) Error() string { return "inventory reservation has expired" }
