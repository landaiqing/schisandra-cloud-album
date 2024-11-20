package model

import "time"

type ScaCommentLikes struct {
	Id        int64     `xorm:"bigint(20) 'id' comment('主键id') pk autoincr notnull " json:"id"`               // 主键id
	TopicId   string    `xorm:"varchar(255) 'topic_id' comment('话题ID') notnull " json:"topic_id"`             // 话题ID
	UserId    string    `xorm:"varchar(255) 'user_id' comment('用户ID') notnull " json:"user_id"`               // 用户ID
	CommentId int64     `xorm:"bigint(20) 'comment_id' comment('评论ID') notnull " json:"comment_id"`           // 评论ID
	LikeTime  time.Time `xorm:"timestamp 'like_time' created comment('点赞时间') default NULL " json:"like_time"` // 点赞时间
}

func (s *ScaCommentLikes) TableName() string {
	return "sca_comment_likes"
}
