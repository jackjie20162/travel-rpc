package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// RegionTranslation 区域（国家/城市）的多语言名称译文。
// Region 行的 name 为基础语言（zh-CN），本表按 locale 存放各语言名称。
type RegionTranslation struct{ ent.Schema }

func (RegionTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("region_id").Comment("关联区域 ID"),
		field.String("locale").NotEmpty().Comment("目标语言：zh-CN/en-US/ar"),
		field.String("name").Optional().Comment("区域名称译文"),
		field.String("source").Default("MANUAL").Comment("来源：MACHINE/MANUAL"),
		field.String("status").Default("DONE").Comment("状态：PENDING/DONE/FAILED"),
		field.Int64("updated_at").Default(0).Comment("最近翻译时间戳(unix秒)"),
	}
}

func (RegionTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("region_id", "locale").Unique(),
		index.Fields("locale", "status"),
	}
}
