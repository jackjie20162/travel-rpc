package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
)

type Traveler struct { ent.Schema }

func (Traveler) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("order_id"),
		field.String("name").NotEmpty(),
		field.String("email").Optional(),
		field.String("phone").Optional(),
		field.String("nationality").Optional(),
	}
}
