package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ScaAuthPermissionRule holds the model definition for the ScaAuthPermissionRule entity.
type ScaAuthPermissionRule struct {
	ent.Schema
}

// Fields of the ScaAuthPermissionRule.
func (ScaAuthPermissionRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20) unsigned",
			}).
			Unique(),
		field.String("ptype").
			MaxLen(100).
			Nillable(),
		field.String("v0").
			MaxLen(100).
			Nillable(),
		field.String("v1").
			MaxLen(100).
			Nillable(),
		field.String("v2").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("v3").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("v4").
			MaxLen(100).
			Optional().
			Nillable(),
		field.String("v5").
			MaxLen(100).
			Optional().
			Nillable().Annotations(
			entsql.WithComments(true),
		),
	}
}

// Edges of the ScaAuthPermissionRule.
func (ScaAuthPermissionRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sca_auth_role", ScaAuthRole.Type).
			Ref("sca_auth_permission_rule").
			Unique(),
	}
}

// Annotations of the ScaAuthPermissionRule.
func (ScaAuthPermissionRule) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("角色权限规则表"),
		entsql.Annotation{
			Table: "sca_auth_permission_rule",
		},
	}
}
