package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ScaAuthPermissionRule holds the model definition for the ScaAuthPermissionRule entity.
type ScaAuthPermissionRule struct {
	ent.Schema
}

// Fields of the ScaAuthPermissionRule.
func (ScaAuthPermissionRule) Fields() []ent.Field {
	return []ent.Field{
		field.Int("id").
			SchemaType(map[string]string{
				dialect.MySQL: "int(11)",
			}).
			Unique(),
		field.String("ptype").
			MaxLen(100).
			Optional(),
		field.String("v0").
			MaxLen(100).
			Optional(),
		field.String("v1").
			MaxLen(100).
			Optional(),
		field.String("v2").
			MaxLen(100).
			Optional().
			Optional(),
		field.String("v3").
			MaxLen(100).
			Optional(),
		field.String("v4").
			MaxLen(100).
			Optional(),
		field.String("v5").
			MaxLen(100).
			Optional().
			Annotations(
				entsql.WithComments(true),
			),
	}
}

// Edges of the ScaAuthPermissionRule.
func (ScaAuthPermissionRule) Edges() []ent.Edge {
	return nil
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
