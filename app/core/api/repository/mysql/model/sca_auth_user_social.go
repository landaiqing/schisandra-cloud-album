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

// ScaAuthUserSocial holds the model definition for the ScaAuthUserSocial entity.
type ScaAuthUserSocial struct {
	ent.Schema
}

func (ScaAuthUserSocial) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DefaultMixin{},
	}
}

// Fields of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Unique().
			Comment("主键ID"),
		field.String("user_id").
			MaxLen(20).
			Comment("用户ID"),
		field.String("open_id").
			MaxLen(50).
			Comment("第三方用户的 open id"),
		field.String("source").
			MaxLen(10).
			Comment("第三方用户来源"),
		field.Int("status").
			Default(0).
			Comment("状态 0正常 1 封禁").Annotations(
			entsql.WithComments(true),
		),
	}
}

// Edges of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
		index.Fields("user_id").
			Unique(),
		index.Fields("open_id").
			Unique(),
	}
}

// Annotations of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户第三方登录信息"),
		entsql.Annotation{
			Table: "sca_auth_user_social",
		},
	}
}
