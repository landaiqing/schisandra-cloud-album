package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

// ScaAuthRole holds the schema definition for the ScaAuthRole entity.
type ScaAuthRole struct {
	ent.Schema
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
			Comment("角色关键字"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("update_at").
			Default(time.Now).UpdateDefault(time.Now).
			Comment("更新时间"),
		field.Int("deleted").
			Default(0).
			Comment("是否删除 0 未删除 1已删除"),
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
