package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Order struct { ent.Schema }
func (Order) Fields() []ent.Field { return []ent.Field{field.Int64("tenant_id"), field.Int64("merchant_id"), field.String("order_no").NotEmpty(), field.Int64("user_id").Optional().Comment("归属人：创建订单的登录用户ID"), field.Int64("customer_id").Optional().Comment("联系人ID"), field.String("customer_email").Optional(), field.String("customer_name").Optional(), field.String("customer_phone").Optional(), field.Int64("total_amount"), field.String("currency").Default("AED"), field.String("status").Default("PENDING_PAYMENT"), field.String("payment_status").Default("PENDING"), field.String("remark").Optional(), field.String("product_name").Optional(), field.String("package_name").Optional(), field.String("service_date").Optional(), field.String("time_slot").Optional()} }
func (Order) Indexes() []ent.Index { return []ent.Index{index.Fields("order_no").Unique(), index.Fields("tenant_id", "merchant_id", "status"), index.Fields("tenant_id", "user_id"), index.Fields("tenant_id", "customer_id")} }
