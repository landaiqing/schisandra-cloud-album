package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// ScaUserLevel holds the model definition for the ScaUserLevel entity.
type ScaUserLevel struct {
	ent.Schema
}

// Fields of the ScaUserLevel.
func (ScaUserLevel) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Comment("主键"),
		field.String("user_id").
			Comment("用户Id"),
		field.Uint8("level_type").
			Comment("等级类型"),
		field.Int("level").
			Comment("等级"),
		field.String("level_name").
			MaxLen(50).
			Comment("等级名称"),
		field.Int64("exp_start").
			Comment("开始经验值"),
		field.Int64("exp_end").
			Comment("结束经验值"),
		field.Text("level_description").
			Optional().
			Comment("等级描述"),
	}
}

// Edges of the ScaUserLevel.
func (ScaUserLevel) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaUserLevel.
func (ScaUserLevel) Indexes() []ent.Index {
	return []ent.Index{}
}

// Annotations of the ScaUserLevel.
func (ScaUserLevel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("用户等级表"),
		entsql.Annotation{
			Table: "sca_user_level",
		},
	}
}
