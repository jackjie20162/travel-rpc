package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ItineraryStop represents a stop/segment in a product itinerary.
// Types: MEETING(集合) / ACTIVITY(活动) / TRANSPORT(交通) / MEAL(餐饮) / RETURN(返程)
type ItineraryStop struct{ ent.Schema }

func (ItineraryStop) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("product_id"),
		field.String("stop_type"), // MEETING / ACTIVITY / TRANSPORT / MEAL / RETURN
		field.String("title").NotEmpty(),
		field.Text("description").Optional(),
		field.Int("sequence").Default(0),

		// 位置信息
		field.String("location_name").Optional(),
		field.String("address").Optional(),
		field.Float("latitude").Optional(),
		field.Float("longitude").Optional(),

		// 高德 POI 信息
		field.String("poi_id").Optional(),        // 高德 POI ID
		field.String("poi_name").Optional(),      // POI 名称

		// 活动特有字段
		field.Bool("is_entering").Default(true),  // 是否入内
		field.String("duration_mode").Optional(), // FIXED / PER_PACKAGE / UNLIMITED
		field.Int("duration_hours").Optional(),   // 体验时长-小时
		field.Int("duration_minutes").Optional(), // 体验时长-分钟
		field.Text("activity_features").Optional(), // 活动体验特色

		// 交通相关
		field.String("transport_type").Optional(), // 车/船/步行等
		field.String("start_time").Optional(),     // HH:mm
		field.String("end_time").Optional(),       // HH:mm

		// 接送地点
		field.String("pickup_location").Optional(),
		field.String("pickup_address").Optional(),
		field.Float("pickup_latitude").Optional(),
		field.Float("pickup_longitude").Optional(),
		field.String("dropoff_location").Optional(),
		field.String("dropoff_address").Optional(),
		field.Float("dropoff_latitude").Optional(),
		field.Float("dropoff_longitude").Optional(),

		field.Bool("is_pickup").Default(false),
		field.Bool("is_dropoff").Default(false),
		field.Bool("agreement_no_shopping").Default(false),
		field.Bool("agreement_adjustable").Default(false),
	}
}

func (ItineraryStop) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "product_id"),
		index.Fields("product_id", "sequence"),
	}
}
