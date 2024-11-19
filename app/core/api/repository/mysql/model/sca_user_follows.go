package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ScaUserFollows holds the model definition for the ScaUserFollows entity.
type ScaUserFollows struct {
	ent.Schema
}

// Fields of the ScaUserFollows.
func (ScaUserFollows) Fields() []ent.Field {
	return []ent.Field{
		field.String("follower_id").
			Comment("关注者"),
		field.String("followee_id").
			Comment("被关注者"),
		field.Uint8("status").
			Default(0).
			Comment("关注状态（0 未互关 1 互关）"),
	}
}

// Edges of the ScaUserFollows.
func (ScaUserFollows) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaUserFollows.
func (ScaUserFollows) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScaUserFollows.
func (ScaUserFollows) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户关注表"),
		entsql.Annotation{
			Table: "sca_user_follows",
		},
	}
}
