package model

import "time"

type ScaCommentReply struct {
	Id              int64     `xorm:"bigint(20) 'id' comment('主键id') pk autoincr notnull " json:"id"`                     // 主键id
	UserId          string    `xorm:"varchar(255) 'user_id' comment('评论用户id') notnull " json:"user_id"`                   // 评论用户id
	TopicId         string    `xorm:"varchar(255) 'topic_id' comment('评论话题id') notnull " json:"topic_id"`                 // 评论话题id
	TopicType       int       `xorm:"bigint(20) 'topic_type' comment('话题类型') notnull " json:"topic_type"`                 // 话题类型
	Content         string    `xorm:"varchar(255) 'content' comment('评论内容') notnull " json:"content"`                     // 评论内容
	CommentType     int       `xorm:"bigint(20) 'comment_type' comment('评论类型 0评论 1 回复') notnull " json:"comment_type"`    // 评论类型 0评论 1 回复
	ReplyTo         int64     `xorm:"bigint(20) 'reply_to' comment('回复子评论ID') default NULL " json:"reply_to"`             // 回复子评论ID
	ReplyId         int64     `xorm:"bigint(20) 'reply_id' comment('回复父评论Id') default NULL " json:"reply_id"`             // 回复父评论Id
	ReplyUser       string    `xorm:"varchar(255) 'reply_user' comment('回复人id') default NULL " json:"reply_user"`         // 回复人id
	Author          int       `xorm:"bigint(20) 'author' comment('评论回复是否作者  0否 1是') notnull default 0 " json:"author"`    // 评论回复是否作者  0否 1是
	Likes           int64     `xorm:"bigint(20) 'likes' comment('点赞数') default 0 " json:"likes"`                          // 点赞数
	ReplyCount      int64     `xorm:"bigint(20) 'reply_count' comment('回复数量') default 0 " json:"reply_count"`             // 回复数量
	Deleted         int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"` // 是否删除 0 未删除 1 已删除
	Browser         string    `xorm:"varchar(255) 'browser' comment('浏览器') notnull " json:"browser"`                      // 浏览器
	OperatingSystem string    `xorm:"varchar(255) 'operating_system' comment('操作系统') notnull " json:"operating_system"`   // 操作系统
	CommentIp       string    `xorm:"varchar(255) 'comment_ip' comment('IP地址') notnull " json:"comment_ip"`               // IP地址
	Location        string    `xorm:"varchar(255) 'location' comment('地址') notnull " json:"location"`                     // 地址
	Agent           string    `xorm:"varchar(255) 'agent' comment('设备信息') notnull " json:"agent"`                         // 设备信息
	CreatedAt       time.Time `xorm:"timestamp 'created_at' created comment('创建时间') default NULL " json:"created_at"`     // 创建时间
	UpdatedAt       time.Time `xorm:"timestamp 'updated_at'  updated comment('更新时间') default NULL " json:"updated_at"`    // 更新时间
	Version         int64     `xorm:"bigint(20) 'version' version comment('版本') default 0 " json:"version"`               // 版本
	DeletedAt       time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`      // 删除时间

}

func (s *ScaCommentReply) TableName() string {
	return "sca_comment_reply"
}
