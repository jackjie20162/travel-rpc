package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Currency is a global dictionary of supported display currencies.
// Prices are stored in the base currency (AED); this table drives the
// selectable currency list plus symbol/decimals/position rendering.
type Currency struct{ ent.Schema }

func (Currency) Fields() []ent.Field {
	return []ent.Field{
		field.String("code").NotEmpty().Comment("ISO 4217 币种代码，如 AED/USD/CNY"),
		field.String("symbol").Optional().Comment("货币符号，如 د.إ/$/¥"),
		field.String("name_key").Optional().Comment("i18n 名称键"),
		field.Int("decimals").Default(2).Comment("小数位数"),
		field.String("symbol_position").Default("prefix").Comment("符号位置：prefix/suffix"),
		field.Bool("is_base").Default(false).Comment("是否基准币"),
		field.String("status").Default("ACTIVE").Comment("状态：ACTIVE/INACTIVE"),
		field.Int("sort").Default(0).Comment("排序权重，越小越靠前"),
	}
}

func (Currency) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("status", "sort"),
	}
}
