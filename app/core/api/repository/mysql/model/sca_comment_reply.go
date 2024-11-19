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

// ScaCommentReply holds the model definition for the ScaCommentReply entity.
type ScaCommentReply struct {
	ent.Schema
}

func (ScaCommentReply) Mixin() []ent.Mixin {
	return []ent.Mixin{
		mixin.DefaultMixin{},
	}
}

// Fields of the ScaCommentReply.
func (ScaCommentReply) Fields() []ent.Field {
	return []ent.Field{
		field.Int64("id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Unique().
			Comment("主键id"),
		field.String("user_id").
			Comment("评论用户id"),
		field.String("topic_id").
			Comment("评论话题id"),
		field.Int("topic_type").
			Comment("话题类型"),
		field.String("content").
			Comment("评论内容"),
		field.Int("comment_type").
			Comment("评论类型 0评论 1 回复"),
		field.Int64("reply_to").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Optional().
			Comment("回复子评论ID"),
		field.Int64("reply_id").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Optional().
			Comment("回复父评论Id"),
		field.String("reply_user").
			Optional().
			Comment("回复人id"),
		field.Int("author").
			Default(0).
			Comment("评论回复是否作者  0否 1是"),
		field.Int64("likes").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Optional().
			Default(0).
			Comment("点赞数"),
		field.Int64("reply_count").
			SchemaType(map[string]string{
				dialect.MySQL: "bigint(20)",
			}).
			Optional().
			Default(0).
			Comment("回复数量"),
		field.String("browser").
			Comment("浏览器"),
		field.String("operating_system").
			Comment("操作系统"),
		field.String("comment_ip").
			Comment("IP地址"),
		field.String("location").
			Comment("地址"),
		field.String("agent").
			MaxLen(255).
			Comment("设备信息"),
	}
}

// Edges of the ScaCommentReply.
func (ScaCommentReply) Edges() []ent.Edge {
	return nil
}

// Indexes of the ScaCommentReply.
func (ScaCommentReply) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("id").
			Unique(),
	}
}

// Annotations of the ScaCommentReply.
func (ScaCommentReply) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.WithComments(true),
		schema.Comment("评论回复表"),
		entsql.Annotation{
			Table: "sca_comment_reply",
		},
	}
}
