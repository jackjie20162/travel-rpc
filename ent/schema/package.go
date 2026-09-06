package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ProductPackage represents a sellable package/SKU under a tourism product.
// The type name intentionally avoids the Go keyword `package` while Ent
// continues to derive the database table from the schema type.
type ProductPackage struct{ ent.Schema }

func (ProductPackage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id"),
		field.Int64("merchant_id"),
		field.Int64("product_id"),
		field.String("code").NotEmpty(),
		field.String("name").NotEmpty(),
		field.String("status").Default("ACTIVE"),
		// --- 套餐发布扩展字段 ---
		field.String("pricing_mode").Default("SAME_PRICE"),   // SAME_PRICE / GROUP_PRICE / TIER_PRICE
		field.String("inventory_mode").Default("UNLIMITED"), // UNLIMITED / DAILY / TOTAL
		field.Int("min_order_qty").Default(1),               // 起订人数
		field.String("sell_currency").Default("AED"),        // 卖价币种
		field.String("cost_currency").Default("AED"),        // 底价币种
		field.Text("group_prices").Optional(),               // 人群报价 JSON
		field.Text("tier_prices").Optional(),                // 阶梯报价 JSON
	}
}

func (ProductPackage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("tenant_id", "product_id", "code").Unique(),
		index.Fields("tenant_id", "product_id", "status"),
	}
}
