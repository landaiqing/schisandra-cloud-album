package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model/mixin"
)

// ScaAuthRole holds the model definition for the ScaAuthRole entity.
type ScaAuthRole struct {
	ent.Schema
}

func (ScaAuthRole) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DefaultMixin{},
	}
}

// Fields of the ScaAuthRole.
func (ScaAuthRole) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Unique().
			Comment("主键ID"),
		field.String("role_name").
			MaxLen(32).
			Comment("角色名称"),
		field.String("role_key").
			MaxLen(64).
			Comment("角色关键字").
			Annotations(
				entsql.WithComments(true),
			),
	}
}

// Edges of the ScaAuthRole.
func (ScaAuthRole) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("sca_auth_permission_rule", ScaAuthPermissionRule.Type),
	}
}

// Indexes of the ScaAuthRole.
func (ScaAuthRole) Indexes() []ent.Index {
	return nil
}

// Annotations of the ScaAuthRole.
func (ScaAuthRole) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("角色表"),
		entsql.Annotation{
			Table: "sca_auth_role",
		},
	}
}
