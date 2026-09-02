package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Voucher struct { ent.Schema }

func (Voucher) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("order_id"),
		field.String("voucher_no").NotEmpty(),
		field.String("status").Default("ISSUED"),
		field.String("redeemed_at").Optional(),
	}
}
