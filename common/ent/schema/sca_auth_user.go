package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScaAuthUser holds the schema definition for the ScaAuthUser entity.
type ScaAuthUser struct {
	ent.Schema
}

// Fields of the ScaAuthUser.
func (ScaAuthUser) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Unique().
			Comment("自增ID"),
		field.String("uid").
			MaxLen(20).
			Unique().
			Comment("唯一ID"),
		field.String("username").
			MaxLen(32).
			Optional().
			Comment("用户名"),
		field.String("nickname").
			MaxLen(32).
			Optional().
			Comment("昵称"),
		field.String("email").
			MaxLen(32).
			Optional().
			Comment("邮箱"),
		field.String("phone").
			MaxLen(32).
			Optional().
			Comment("电话"),
		field.String("password").
			MaxLen(64).
			Optional().
			Sensitive().
			Comment("密码"),
		field.String("gender").
			MaxLen(32).
			Optional().
			Comment("性别"),
		field.String("avatar").
			Optional().
			Comment("头像"),
		field.Int8("status").
			Default(0).
			Optional().
			Comment("状态 0 正常 1 封禁"),
		field.String("introduce").
			MaxLen(255).
			Optional().
			Comment("介绍"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("update_at").
			Default(time.Now).UpdateDefault(time.Now).
			Nillable().
			Optional().
			Comment("更新时间"),
		field.Int8("deleted").
			Default(0).
			Comment("是否删除 0 未删除 1 已删除"),
		field.String("blog").
			MaxLen(30).
			Nillable().
			Optional().
			Comment("博客"),
		field.String("location").
			MaxLen(50).
			Nillable().
			Optional().
			Comment("地址"),
		field.String("company").
			MaxLen(50).
			Nillable().
			Optional().
			Comment("公司"),
	}
}

// Edges of the ScaAuthUser.
func (ScaAuthUser) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sca_auth_user_social", ScaAuthUserSocial.Type),
		edge.To("sca_auth_user_device", ScaAuthUserDevice.Type),
	}
}

// Indexes of the ScaAuthUser.
func (ScaAuthUser) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
		index.Fields("uid").
			Unique(),
		index.Fields("phone").
			Unique(),
	}
}
