package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

func (Traveler) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "order_id"),
	}
}
