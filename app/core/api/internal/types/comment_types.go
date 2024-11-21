package types

import "time"

// CommentResponse 评论响应
type CommentResponse struct {
	Id              int64     `json:"id"`
	Content         string    `json:"content"`
	UserId          string    `json:"user_id"`
	TopicId         string    `json:"topic_id"`
	Author          int64     `json:"author"`
	Location        string    `json:"location"`
	Browser         string    `json:"browser"`
	OperatingSystem string    `json:"operating_system"`
	CreatedTime     time.Time `json:"created_time"`
	ReplyId         int64     `json:"reply_id,omitempty"`
	ReplyUser       string    `json:"reply_user,omitempty"`
	ReplyTo         int64     `json:"reply_to,omitempty"`
}

// CommentImages 评论图片
type CommentImages struct {
	TopicId   string   `json:"topic_id" bson:"topic_id"`
	CommentId int64    `json:"comment_id" bson:"comment_id"`
	UserId    string   `json:"user_id" bson:"user_id"`
	Images    [][]byte `json:"images" bson:"images"`
	CreatedAt string   `json:"created_at" bson:"created_at"`
}
