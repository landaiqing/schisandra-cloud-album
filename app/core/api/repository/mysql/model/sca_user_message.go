package model

import "time"

type ScaUserMessage struct {
	Id          int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                           // 主键
	TopicId     string    `xorm:"varchar(255) 'topic_id' comment('话题Id') notnull " json:"topic_id"`                       // 话题Id
	FromId      string    `xorm:"varchar(255) 'from_id' comment('来自人') notnull " json:"from_id"`                          // 来自人
	ToId        string    `xorm:"varchar(255) 'to_id' comment('送达人') notnull " json:"to_id"`                              // 送达人
	Content     string    `xorm:"varchar(255) 'content' comment('消息内容') notnull " json:"content"`                         // 消息内容
	IsRead      int64     `xorm:"bigint(20) 'is_read' comment('是否已读') default NULL " json:"is_read"`                      // 是否已读
	CreatedBy   string    `xorm:"varchar(32) 'created_by' comment('创建人') default NULL " json:"created_by"`                // 创建人
	CreatedTime time.Time `xorm:"datetime 'created_time' comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdateBy    string    `xorm:"varchar(32) 'update_by' comment('更新人') default NULL " json:"update_by"`                  // 更新人
	UpdateTime  time.Time `xorm:"datetime 'update_time' comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"`   // 更新时间
	Deleted     int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"`     // 是否删除 0 未删除 1 已删除
	CreatedAt   time.Time `xorm:"timestamp 'created_at' created comment('创建时间') default NULL " json:"created_at"`         // 创建时间
	UpdatedAt   time.Time `xorm:"timestamp 'updated_at' updated comment('更新时间') default NULL " json:"updated_at"`         // 更新时间
	DeletedAt   time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`          // 删除时间
}

func (s *ScaUserMessage) TableName() string {
	return "sca_user_message"
}
