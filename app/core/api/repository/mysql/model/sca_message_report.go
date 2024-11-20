package model

import "time"

type ScaMessageReport struct {
	Id            int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                                // 主键
	UserId        string    `xorm:"varchar(20) 'user_id' comment('用户Id') default NULL " json:"user_id"`                          // 用户Id
	Type          int32     `xorm:"int(11) 'type' comment('举报类型 0评论 1 相册') default NULL " json:"type"`                           // 举报类型 0评论 1 相册
	CommentId     int64     `xorm:"bigint(20) 'comment_id' comment('评论Id') default NULL " json:"comment_id"`                     // 评论Id
	TopicId       string    `xorm:"varchar(20) 'topic_id' comment('话题Id') default NULL " json:"topic_id"`                        // 话题Id
	ReportType    string    `xorm:"varchar(255) 'report_type' comment('举报') default NULL " json:"report_type"`                   // 举报
	ReportContent string    `xorm:"longtext 'report_content' comment('举报说明内容') " json:"report_content"`                          // 举报说明内容
	ReportTag     string    `xorm:"varchar(255) 'report_tag' comment('举报标签') default NULL " json:"report_tag"`                   // 举报标签
	Status        int32     `xorm:"int(11) 'status' comment('状态（0 未处理 1 已处理）') default NULL " json:"status"`                     // 状态（0 未处理 1 已处理）
	CreatedAt     time.Time `xorm:"datetime 'created_at' comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"`        // 创建时间
	UpdateBy      string    `xorm:"varchar(32) 'update_by' created comment('更新人') default NULL " json:"update_by"`               // 更新人
	UpdatedAt     time.Time `xorm:"datetime 'updated_at' updated comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"` // 更新时间
	Deleted       int32     `xorm:"int(11) 'deleted' comment('是否删除 0否 1是') default 0 " json:"deleted"`                           // 是否删除 0否 1是
	DeletedAt     time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`               // 删除时间

}

func (s *ScaMessageReport) TableName() string {
	return "sca_message_report"
}
