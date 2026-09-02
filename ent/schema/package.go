package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductPackage represents a sellable package/SKU under a tourism product.
// The type name intentionally avoids the Go keyword `package` while Ent
// continues to derive the database table from the schema type.
type ProductPackage struct{ ent.Schema }

func (ProductPackage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("product_id"),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("status").Default("ACTIVE"),
	}
}

func (ProductPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "product_id", "code").Unique(),
		index.Fields("tenant_id", "product_id", "status"),
	}
}
