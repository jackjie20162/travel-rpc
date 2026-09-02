package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Merchant struct { ent.Schema }

func (Merchant) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.String("name").NotEmpty(),
		field.String("code").NotEmpty(),
		field.String("status").Default("ACTIVE"),
	}
}
