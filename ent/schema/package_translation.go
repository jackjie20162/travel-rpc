package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// PackageTranslation holds translated copies of a product package's name for a
// given locale. The ProductPackage row remains the base-language source of
// truth; this table is derived content for the C-end catalog.
type PackageTranslation struct{ ent.Schema }

func (PackageTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("package_id").Comment("关联套餐 ID"),
		field.String("locale").NotEmpty().Comment("目标语言：zh-CN/en-US/ar"),
		field.String("name").Optional().Comment("套餐名称译文"),
		field.String("source").Default("MACHINE").Comment("来源：MACHINE/MANUAL"),
		field.String("status").Default("PENDING").Comment("状态：PENDING/DONE/FAILED"),
		field.Int64("updated_at").Default(0).Comment("最近翻译时间戳(unix秒)"),
	}
}

func (PackageTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("package_id", "locale").Unique(),
		index.Fields("locale", "status"),
	}
}
