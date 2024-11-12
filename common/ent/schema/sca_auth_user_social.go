package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScaAuthUserSocial holds the schema definition for the ScaAuthUserSocial entity.
type ScaAuthUserSocial struct {
	ent.Schema
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
			Comment("状态 0正常 1 封禁"),
		field.Time("created_at").
			Default(time.Now).
			Immutable().
			Comment("创建时间"),
		field.Time("update_at").
			Default(time.Now).UpdateDefault(time.Now).
			Optional().
			Nillable().
			Comment("更新时间"),
		field.Int("deleted").
			Default(0).
			Comment("是否删除"),
	}
}

// Edges of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sca_auth_user", ScaAuthUser.Type).
			Ref("sca_auth_user_social").
			Unique(),
	}
}

// Indexes of the ScaAuthUserSocial.
func (ScaAuthUserSocial) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id", "user_id", "open_id").
			Unique(),
	}
}
