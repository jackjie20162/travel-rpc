package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// User holds the schema definition for the end-user (customer) account.
// The fields mirror the legacy PHP user table so that existing data can be
// migrated without transformation.
type User struct{ ent.Schema }

func (User) Fields() []ent.Field {
	return []ent.Field{
		field.String("username").Default(""),
		field.String("nickname").Default(""),
		field.String("password").Default(""),
		field.String("salt").Default(""),
		field.String("email").Default(""),
		field.String("mobile").Default(""),
		field.String("avatar").Default(""),
		field.Uint8("level").Default(0),
		field.Int8("gender").Default(0),
		field.String("birthday").Optional().Nillable(),
		field.String("bio").Default(""),
		field.Float("money").Default(0),
		field.Int("score").Default(0),
		field.Uint32("successions").Default(1),
		field.Uint32("maxsuccessions").Default(1),
		field.Int64("prevtime").Default(0),
		field.Int64("logintime").Default(0),
		field.String("loginip").Default(""),
		field.Uint8("loginfailure").Default(0),
		field.Int64("loginfailuretime").Optional().Nillable(),
		field.String("joinip").Default(""),
		field.Int64("jointime").Optional().Nillable(),
		field.Int64("createtime").Optional().Nillable(),
		field.Int64("updatetime").Optional().Nillable(),
		field.String("token").Default(""),
		field.String("status").Default("normal"),
		field.String("verification").Default(""),
	}
}

func (User) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("username").Unique(),
		index.Fields("email"),
		index.Fields("mobile"),
		index.Fields("token"),
		index.Fields("status"),
	}
}
