package model

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
)

// ScaCommentLikes holds the model definition for the ScaCommentLikes entity.
type ScaCommentLikes struct {
	ent.Schema
}

// Fields of the ScaCommentLikes.
func (ScaCommentLikes) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Comment("主键id"),
		field.String("topic_id").
			Comment("话题ID"),
		field.String("user_id").
			Comment("用户ID"),
		field.Int64("comment_id").
			Comment("评论ID"),
		field.Time("like_time").
			Default(time.Now).
			Comment("点赞时间"),
	}
}

// Edges of the ScaCommentLikes.
func (ScaCommentLikes) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaCommentLikes.
func (ScaCommentLikes) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id"),
		index.Fields("comment_id"),
	}
}

// Annotations of the ScaCommentLikes.
func (ScaCommentLikes) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("评论点赞表"),
		entsql.Annotation{
			Table: "sca_comment_likes",
		},
	}
}
