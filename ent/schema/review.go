package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Review struct{ ent.Schema }

func (Review) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("order_id"),
		field.String("order_no").NotEmpty(),
		field.Int64("product_id"),
		field.Int64("user_id").Optional().Comment("评价人（客户端用户ID）"),
		field.Int("rating").Default(5).Comment("总评 1-5"),
		field.Int("service_rating").Optional().Comment("服务评分 1-5"),
		field.Int("value_rating").Optional().Comment("性价比评分 1-5"),
		field.String("content").Optional().Comment("评价内容"),
		field.String("images").Optional().Comment("评价图片JSON数组"),
		field.String("reply_content").Optional().Comment("商户回复"),
		field.Int64("reply_time").Optional().Comment("商户回复时间"),
		field.String("status").Default("PENDING").Comment("PENDING/APPROVED/REPLIED"),
	}
}

func (Review) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		index.Fields("product_id", "status"),
		index.Fields("tenant_id", "merchant_id", "status"),
		index.Fields("user_id"),
	}
}
