package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

type Order struct{ ent.Schema }

func (Order) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.String("order_no").NotEmpty(),
		field.Int64("user_id").Optional().Comment("归属人：创建订单的登录用户ID"),
		field.Int64("customer_id").Optional().Comment("联系人ID"),
		field.String("customer_email").Optional(),
		field.String("customer_name").Optional(),
		field.String("customer_phone").Optional(),
		field.Int64("total_amount"),
		field.String("currency").Default("AED"),
		// --- 下单锁汇字段：结算真值仍为 total_amount(基准币 AED)，以下仅用于用户侧展示 ---
		field.String("display_currency").Default("AED").Comment("下单时用户选择的展示币种"),
		field.Int64("exchange_rate_micro").Default(1000000).Comment("下单时锁定的汇率×1e6，基准币为1000000"),
		field.Int64("display_amount_minor").Default(0).Comment("锁定的展示金额(最小货币单位,×100)"),
		field.String("status").Default("PENDING_PAYMENT"),
		field.String("payment_status").Default("PENDING"),
		field.String("remark").Optional(),
		field.String("product_name").Optional(),
		field.String("package_name").Optional(),
		field.String("service_date").Optional(),
		field.String("time_slot").Optional(),
		field.String("reject_reason").Optional().Comment("商户拒绝接单原因"),
		field.Int64("verified_at").Optional().Comment("核销时间戳"),
		field.String("prev_status").Optional().Comment("退款申请前的状态，用于拒绝退款时恢复"),
	}
}

func (Order) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("order_no").Unique(),
		index.Fields("tenant_id", "merchant_id", "status"),
		index.Fields("tenant_id", "user_id"),
		index.Fields("tenant_id", "customer_id"),
	}
}
