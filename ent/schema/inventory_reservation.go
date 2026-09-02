package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// InventoryReservation records an idempotent inventory hold.
// The reservation lifecycle is RESERVED -> CONFIRMED/RELEASED/EXPIRED.
type InventoryReservation struct { ent.Schema }

func (InventoryReservation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("inventory_id"),
		field.String("reservation_key").NotEmpty(),
		field.Int("quantity").Positive(),
		field.Int64("order_id").Optional(),
		field.String("status").Default("RESERVED"),
		field.Time("expires_at").Default(func() time.Time { return time.Now().Add(15 * time.Minute) }),
		field.Time("created_at").Default(time.Now),
		field.Time("updated_at").Default(time.Now).UpdateDefault(time.Now),
	}
}

func (InventoryReservation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "reservation_key").Unique(),
		index.Fields("inventory_id", "status"),
		index.Fields("tenant_id", "order_id"),
		index.Fields("status", "expires_at"),
	}
}
