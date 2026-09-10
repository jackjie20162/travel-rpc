package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ExchangeRate stores the conversion rate from the base currency to a target
// currency. rate_micro holds the real rate multiplied by 1e6 as an int64 to
// avoid float precision drift. Queries pick the latest ACTIVE record per pair
// (ordered by effective_at desc) as the effective rate.
type ExchangeRate struct{ ent.Schema }

func (ExchangeRate) Fields() []ent.Field {
	return []ent.Field{
		field.String("base_currency").NotEmpty().Default("AED").Comment("基准币种"),
		field.String("target_currency").NotEmpty().Comment("目标币种"),
		field.Int64("rate_micro").Default(1000000).Comment("实际汇率×1e6"),
		field.String("source").Default("MANUAL").Comment("来源：MANUAL/API"),
		field.Int64("effective_at").Default(0).Comment("生效时间戳(unix秒)"),
		field.String("status").Default("ACTIVE").Comment("状态：ACTIVE/INACTIVE"),
	}
}

func (ExchangeRate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("base_currency", "target_currency", "effective_at").Unique(),
		index.Fields("base_currency", "target_currency", "status"),
	}
}
