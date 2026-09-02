package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Product struct { ent.Schema }

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"), field.Int64("merchant_id"), field.String("code").NotEmpty(), field.String("title").NotEmpty(), field.String("slug").Optional(), field.String("destination").Optional(), field.Text("description").Optional(), field.String("currency").Default("AED"), field.Int64("min_price").Default(0), field.String("status").Default("DRAFT"),
	}
}
func (Product) Indexes() []ent.Index {
	return []ent.Index{index.Fields("tenant_id", "code").Unique(), index.Fields("tenant_id", "destination", "status")}
}
