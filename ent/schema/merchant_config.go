package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// MerchantConfig stores merchant-level key-value configurations.
// Config types: email (邮件SMTP配置), map (高德地图配置)
type MerchantConfig struct{ ent.Schema }

func (MerchantConfig) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("tenant_id").Default(0),
		field.Int64("merchant_id").Default(0),
		field.String("config_type").NotEmpty(), // email / map
		field.String("config_key").NotEmpty(),  // 配置项键名
		field.Text("config_value").Optional(),  // 配置项值（JSON 或纯文本）
		field.String("description").Optional(), // 配置项说明
	}
}

func (MerchantConfig) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("merchant_id", "config_type"),
		index.Fields("merchant_id", "config_type", "config_key").Unique(),
	}
}
