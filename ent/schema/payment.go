package schema

import (
    "entgo.io/ent"
    "entgo.io/ent/schema/field"
    "entgo.io/ent/schema/index"
)

type Payment struct { ent.Schema }

func (Payment) Fields() []ent.Field {
    return []ent.Field{
        field.Int64("tenant_id"),
        field.Int64("merchant_id"),
        field.Int64("order_id"),
        field.String("payment_no").NotEmpty(),
        field.String("provider").NotEmpty(),
        field.String("provider_payment_id").Optional(),
        field.Int64("amount"),
        field.String("currency").Default("AED"),
        field.String("status").Default("CREATED"),
        field.String("idempotency_key").NotEmpty(),
    }
}

func (Payment) Indexes() []ent.Index {
    return []ent.Index{
        index.Fields("payment_no").Unique(),
        index.Fields("tenant_id", "merchant_id", "order_id"),
        index.Fields("tenant_id", "idempotency_key").Unique(),
        index.Fields("provider", "provider_payment_id").Unique(),
        index.Fields("status"),
    }
}
