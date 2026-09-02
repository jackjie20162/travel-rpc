package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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

func (Merchant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "code").Unique(),
		index.Fields("tenant_id", "status"),
	}
}
