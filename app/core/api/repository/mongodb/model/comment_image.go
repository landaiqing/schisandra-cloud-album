package model

import "github.com/chenmingyong0423/go-mongox/v2"

type CommentImage struct {
	mongox.Model `bson:",inline"`
	TopicID      string   `bson:"topic_id"`
	CommentID    string   `bson:"comment_id"`
	UserID       string   `bson:"user_id"`
	Images       []string `bson:"images"`
}
