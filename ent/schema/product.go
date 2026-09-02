package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Product struct { ent.Schema }

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.String("code").NotEmpty(),
		field.String("title").NotEmpty(),
		field.String("slug").Optional(),
		field.String("destination").Optional(),
		field.Text("description").Optional(),
		field.String("currency").Default("AED"),
		field.Int64("min_price").Default(0),
		field.String("status").Default("DRAFT"),
	}
}
