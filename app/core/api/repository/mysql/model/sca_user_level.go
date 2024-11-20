package model

import "time"

type ScaUserLevel struct {
	Id          int64     `xorm:"bigint(20) 'id' comment('主键') pk notnull " json:"id"`                                          // 主键
	UserId      string    `xorm:"varchar(255) 'user_id' comment('用户Id') notnull " json:"user_id"`                               // 用户Id
	LevelType   uint8     `xorm:"tinyint(3) UNSIGNED 'level_type' comment('等级类型') notnull " json:"level_type"`                  // 等级类型
	Level       int64     `xorm:"bigint(20) 'level' comment('等级') notnull " json:"level"`                                       // 等级
	LevelName   string    `xorm:"varchar(50) 'level_name' comment('等级名称') notnull " json:"level_name"`                          // 等级名称
	ExpStart    int64     `xorm:"bigint(20) 'exp_start' comment('开始经验值') notnull " json:"exp_start"`                            // 开始经验值
	ExpEnd      int64     `xorm:"bigint(20) 'exp_end' comment('结束经验值') notnull " json:"exp_end"`                                // 结束经验值
	Description string    `xorm:"longtext 'description' comment('等级描述') " json:"description"`                                   // 等级描述
	CreatedAt   time.Time `xorm:"datetime 'created_at' created comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdatedAt   time.Time `xorm:"datetime 'updated_at' updated comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"`  // 更新时间
	DeletedAt   time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`                // 删除时间

}

func (s *ScaUserLevel) TableName() string {
	return "sca_user_level"
}
