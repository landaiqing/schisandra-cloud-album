package types

import (
	"time"

	"github.com/chenmingyong0423/go-mongox/v2"
)

// CommentImages 评论 图片
type CommentImages struct {
	mongox.Model `bson:",inline"`
	TopicId      string   `json:"topic_id" bson:"topic_id"`
	CommentId    int64    `json:"comment_id" bson:"comment_id"`
	UserId       string   `json:"user_id" bson:"user_id"`
	Images       [][]byte `json:"images" bson:"images"`
}

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

// CommentContent 评论内容
type CommentContent struct {
	NickName        string    `json:"nickname"`
	Avatar          string    `json:"avatar"`
	Level           int64     `json:"level,omitempty" default:"0"`
	Id              int64     `json:"id"`
	UserId          string    `json:"user_id"`
	TopicId         string    `json:"topic_id"`
	Content         string    `json:"content"`
	ReplyTo         int64     `json:"reply_to,omitempty"`
	ReplyId         int64     `json:"reply_id,omitempty"`
	ReplyUser       string    `json:"reply_user,omitempty"`
	ReplyNickname   string    `json:"reply_nickname,omitempty"`
	IsAuthor        int64     `json:"is_author"`
	Likes           int64     `json:"likes"`
	ReplyCount      int64     `json:"reply_count"`
	CreatedTime     time.Time `json:"created_time"`
	Location        string    `json:"location"`
	Browser         string    `json:"browser"`
	OperatingSystem string    `json:"operating_system"`
	IsLiked         bool      `json:"is_liked" default:"false"`
	Images          []string  `json:"images,omitempty"`
}

// CommentListPageResponse 评论返回值
type CommentListPageResponse struct {
	Size     int              `json:"size"`
	Total    int64            `json:"total"`
	Current  int              `json:"current"`
	Comments []CommentContent `json:"comments"`
}

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
}
