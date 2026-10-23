package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Attraction 景点库：商户可复用的景点/地点资料，旅游路线（产品）的行程节点可绑定这些景点。
// 坐标统一以 WGS84 存储（latitude/longitude），前端地图选点组件负责坐标系转换。
type Attraction struct{ ent.Schema }

func (Attraction) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.String("name").NotEmpty().Comment("景点名称"),
		field.String("name_en").Optional().Comment("景点英文名称"),
		field.String("category").Optional().Comment("分类: NATURE/HISTORY/THEME_PARK/MUSEUM/RELIGION/SHOPPING/OTHER"),
		field.Text("tags").Optional().Comment("标签 JSON 数组: [\"亲子\",\"网红打卡\"]"),
		field.String("city").Optional().Comment("所在城市（冗余展示名，历史兜底）"),
		field.Int64("city_id").Default(0).Comment("关联城市 region.id，指向 regions 表 level=CITY 的行"),
		field.String("address").Optional().Comment("详细地址"),

		// 地图坐标（WGS84）
		field.Float("latitude").Optional().Comment("纬度 WGS84"),
		field.Float("longitude").Optional().Comment("经度 WGS84"),

		// 高德 POI 信息
		field.String("poi_id").Optional().Comment("高德 POI ID"),
		field.String("poi_name").Optional().Comment("高德 POI 名称"),

		// 图片与简介
		field.String("cover_image").Optional().Comment("封面图 URL"),
		field.Text("images").Optional().Comment("图片集 JSON 数组"),
		field.Text("description").Optional().Comment("简介"),
		field.Text("rich_content").Optional().Comment("图文介绍（富文本 HTML）"),

		// 开放与游玩信息
		field.String("opening_hours").Optional().Comment("开放时间描述"),
		field.String("suggested_duration").Optional().Comment("建议游玩时长，如 2-3小时"),
		field.Text("ticket_info").Optional().Comment("门票参考信息"),

		field.Int("sort").Default(0).Comment("排序，越小越靠前"),
		field.String("status").Default("ACTIVE").Comment("状态: ACTIVE/INACTIVE"),
	}
}

func (Attraction) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "merchant_id", "status"),
		index.Fields("merchant_id", "category"),
	}
}
