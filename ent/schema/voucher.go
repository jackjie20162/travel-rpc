package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Voucher struct { ent.Schema }
func (Voucher) Fields() []ent.Field { return []ent.Field{field.Int64("tenant_id"), field.Int64("order_id"), field.String("voucher_no").NotEmpty(), field.String("status").Default("ISSUED"), field.String("redeemed_at").Optional()} }
func (Voucher) Indexes() []ent.Index { return []ent.Index{index.Fields("voucher_no").Unique(), index.Fields("tenant_id", "order_id"), index.Fields("tenant_id", "status")} }
