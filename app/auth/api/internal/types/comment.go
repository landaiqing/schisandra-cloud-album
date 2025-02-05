package types

import (
	"time"
)

// CommentListQueryResult 评论列表查询结果
type CommentListQueryResult struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	TopicID         string    `json:"topic_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	Author          int64     `json:"author"`
	Likes           int64     `json:"likes"`
	ReplyCount      int64     `json:"reply_count"`
	Browser         string    `json:"browser"`
	OperatingSystem string    `json:"operating_system"`
	Location        string    `json:"location"`
	Avatar          string    `json:"avatar"`
	Nickname        string    `json:"nickname"`
	ImagePath       string    `json:"image_path"`
}

// ReplyListQueryResult 回复列表查询结果
type ReplyListQueryResult struct {
	ID              int64     `json:"id"`
	UserID          string    `json:"user_id"`
	TopicID         string    `json:"topic_id"`
	Content         string    `json:"content"`
	CreatedAt       time.Time `json:"created_at"`
	Author          int64     `json:"author"`
	Likes           int64     `json:"likes"`
	ReplyCount      int64     `json:"reply_count"`
	Browser         string    `json:"browser"`
	OperatingSystem string    `json:"operating_system"`
	Location        string    `json:"location"`
	Avatar          string    `json:"avatar"`
	Nickname        string    `json:"nickname"`
	ReplyUser       string    `json:"reply_user"`
	ReplyId         int64     `json:"reply_id"`
	ReplyTo         int64     `json:"reply_to"`
	ReplyNickname   string    `json:"reply_nickname"`
	ImagePath       string    `json:"image_path"`
}
