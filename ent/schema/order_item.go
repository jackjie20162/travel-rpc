package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type OrderItem struct { ent.Schema }

func (OrderItem) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("order_id"),
		field.Int64("product_id"),
		field.Int64("package_id"),
		field.Int("quantity"),
		field.Int64("unit_price"),
		field.Int64("total_amount"),
		field.String("service_date"),
		field.String("time_slot").Optional(),
	}
}
