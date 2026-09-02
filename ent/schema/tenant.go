package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Tenant struct { ent.Schema }

func (Tenant) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("status").Default("ACTIVE"),
	}
}

func (Tenant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("status"),
	}
}
