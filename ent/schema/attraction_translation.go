package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// AttractionTranslation 景点的可翻译文本按语言的译文。
// Attraction 行的 name/category/tags/address/description/opening_hours/suggested_duration/ticket_info
// 为基础语言（zh-CN），本表按 locale 存放各语言对应字段译文（en-US/ar）。
// 分类译文存放本地化显示名文本；城市名不走本表（由地区字典 region_translations 按 city_id 解析）。
type AttractionTranslation struct{ ent.Schema }

func (AttractionTranslation) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("attraction_id").Comment("关联景点 ID"),
		field.String("locale").NotEmpty().Comment("目标语言：zh-CN/en-US/ar"),
		field.String("name").Optional().Comment("景点名称译文"),
		field.String("category").Optional().Comment("分类译文（本地化显示名文本）"),
		field.Text("tags").Optional().Comment("标签译文（JSON 数组）"),
		field.String("address").Optional().Comment("详细地址译文"),
		field.Text("description").Optional().Comment("简介译文"),
		field.String("opening_hours").Optional().Comment("开放时间译文"),
		field.String("suggested_duration").Optional().Comment("建议游玩时长译文"),
		field.Text("ticket_info").Optional().Comment("门票信息译文"),
		field.String("source").Default("MANUAL").Comment("来源：MACHINE/MANUAL"),
		field.String("status").Default("DONE").Comment("状态：PENDING/DONE/FAILED"),
		field.Int64("updated_at").Default(0).Comment("最近翻译时间戳(unix秒)"),
	}
}

func (AttractionTranslation) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("attraction_id", "locale").Unique(),
		index.Fields("locale", "status"),
	}
}
