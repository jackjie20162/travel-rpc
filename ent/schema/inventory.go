package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Inventory struct { ent.Schema }

func (Inventory) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("package_id"),
		field.String("service_date"),
		field.String("time_slot").Optional(),
		field.Int("capacity").Default(0),
		field.Int("reserved").Default(0),
		field.Int64("unit_price").Default(0),
		field.String("currency").Default("AED"),
		field.String("status").Default("OPEN"),
	}
}
