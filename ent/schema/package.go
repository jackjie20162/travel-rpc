package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Package struct { ent.Schema }
func (Package) Fields() []ent.Field { return []ent.Field{field.Int64("tenant_id"), field.Int64("merchant_id"), field.Int64("product_id"), field.String("code").NotEmpty(), field.String("name").NotEmpty(), field.String("status").Default("ACTIVE")} }
func (Package) Indexes() []ent.Index { return []ent.Index{index.Fields("tenant_id", "product_id", "code").Unique(), index.Fields("tenant_id", "product_id", "status")} }
