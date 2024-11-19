package model

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model/mixin"
)

// ScaCommentMessage holds the model definition for the ScaCommentMessage entity.
type ScaCommentMessage struct {
	ent.Schema
}

func (ScaCommentMessage) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DefaultMixin{},
	}
}

// Fields of the ScaCommentMessage.
func (ScaCommentMessage) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			Unique().
			Comment("主键"),
		field.String("topic_id").
			Comment("话题Id"),
		field.String("from_id").
			Comment("来自人"),
		field.String("to_id").
			Comment("送达人"),
		field.String("content").
			Comment("消息内容"),
		field.Int("is_read").
			Optional().
			Comment("是否已读"),
	}
}

// Edges of the ScaCommentMessage.
func (ScaCommentMessage) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaCommentMessage.
func (ScaCommentMessage) Indexes() []ent.Index {
	return nil
}

// Annotations of the ScaCommentMessage.
func (ScaCommentMessage) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("评论消息表"),
		entsql.Annotation{
			Table: "sca_comment_message",
		},
	}
}
