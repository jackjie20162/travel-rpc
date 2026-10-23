package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// Region 旅游区域字典：国家 -> 城市 两级，单表自关联。
// level=COUNTRY 时 parent_id=0；level=CITY 时 parent_id 指向所属国家。
// 多语言名称存放在 RegionTranslation，本表 name 字段为基础语言（默认 zh-CN）。
// 城市需维护代表 POI 与中心经纬度，景点 attractions.city_id 绑定本表 id。
type Region struct{ ent.Schema }

func (Region) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("parent_id").Default(0).Comment("父级 ID：0 表示国家，城市指向所属国家"),
		field.String("level").NotEmpty().Comment("层级：COUNTRY/CITY"),
		field.String("code").Optional().Comment("ISO 代码，如国家 AE/CN；城市可留空"),
		field.String("name").NotEmpty().Comment("基础名称（默认语言 zh-CN），其余语言见 region_translation"),
		field.String("poi_id").Optional().Comment("代表 POI ID（高德/Google）"),
		field.String("poi_name").Optional().Comment("代表 POI 名称"),
		field.Float("latitude").Optional().Comment("中心纬度 WGS84"),
		field.Float("longitude").Optional().Comment("中心经度 WGS84"),
		field.Int("sort").Default(0).Comment("排序，越小越靠前"),
		field.String("status").Default("ACTIVE").Comment("状态：ACTIVE/INACTIVE"),
	}
}

func (Region) Indexes() []ent.Index {
	return []ent.Index{
		// 同一父级下名称唯一，避免同国家内重复城市名
		index.Fields("parent_id", "name").Unique(),
		index.Fields("level", "status"),
	}
}
