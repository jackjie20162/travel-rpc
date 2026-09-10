package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductTranslation holds machine/human translated copies of a product's
// text fields for a given locale. The product row remains the base-language
// source of truth; this table is derived content used by the C-end catalog
// when the request locale differs from the base language.
type ProductTranslation struct{ ent.Schema }

func (ProductTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("product_id").Comment("关联产品 ID"),
		field.String("locale").NotEmpty().Comment("目标语言：zh-CN/en-US/ar"),
		field.String("title").Optional().Comment("标题译文"),
		field.Text("description").Optional().Comment("描述译文"),
		field.Text("highlights").Optional().Comment("亮点译文(JSON)"),
		field.Text("rich_content").Optional().Comment("图文介绍译文(富文本)"),
		field.Text("booking_notice").Optional().Comment("预订须知译文"),
		field.String("source").Default("MACHINE").Comment("来源：MACHINE/MANUAL"),
		field.String("status").Default("PENDING").Comment("状态：PENDING/DONE/FAILED"),
		field.Int64("updated_at").Default(0).Comment("最近翻译时间戳(unix秒)"),
	}
}

func (ProductTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("product_id", "locale").Unique(),
		index.Fields("locale", "status"),
	}
}
