package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Inventory struct { ent.Schema }
func (Inventory) Fields() []ent.Field { return []ent.Field{field.Int64("tenant_id"), field.Int64("merchant_id"), field.Int64("package_id"), field.String("service_date"), field.String("time_slot").Optional(), field.Int("capacity").Default(0), field.Int("reserved").Default(0), field.Int64("unit_price").Default(0), field.String("currency").Default("AED"), field.String("status").Default("OPEN")} }
func (Inventory) Indexes() []ent.Index { return []ent.Index{index.Fields("package_id", "service_date", "time_slot").Unique(), index.Fields("tenant_id", "package_id", "service_date"), index.Fields("status", "service_date")} }
