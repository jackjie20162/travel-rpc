package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Product struct { ent.Schema }

func (Product) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.String("code").NotEmpty(),
		field.String("title").NotEmpty(),
		field.String("slug").Optional(),
		field.String("destination").Optional(),
		field.Text("description").Optional(),
		field.String("currency").Default("AED"),
		field.Int64("min_price").Default(0),
		field.String("status").Default("DRAFT"),
		// --- 产品发布扩展字段 ---
		field.Text("highlights").Optional(),           // JSON: [{"text":"..."}]
		field.String("cover_image").Optional(),         // 封面图 URL
		field.Text("images").Optional(),                // JSON: ["url1","url2",...]
		field.String("video_url").Optional(),           // 视频 URL
		field.Text("rich_content").Optional(),          // 图文介绍（富文本 HTML）
		field.Text("booking_notice").Optional(),        // 预订须知
	}
}
func (Product) Indexes() []ent.Index {
	return []ent.Index{index.Fields("tenant_id", "code").Unique(), index.Fields("tenant_id", "destination", "status")}
}
