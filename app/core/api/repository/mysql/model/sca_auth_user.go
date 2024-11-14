package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model/mixin"
)

// ScaAuthUser holds the model definition for the ScaAuthUser entity.
type ScaAuthUser struct {
	ent.Schema
}

func (ScaAuthUser) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DefaultMixin{},
	}
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
		field.Int8("gender").
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
			Comment("公司").
			Annotations(
				entsql.WithComments(true),
			),
	}
}

// Edges of the ScaAuthUser.
func (ScaAuthUser) Edges() []ent.Edge {
	return nil
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

// Annotations of the ScaAuthUser.
func (ScaAuthUser) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户表"),
		entsql.Annotation{
			Table: "sca_auth_user",
		},
	}
}
