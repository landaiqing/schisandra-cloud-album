package model

import "time"

type ScaAuthRole struct {
	Id        int64     `xorm:"bigint(20) 'id' comment('主键ID') pk autoincr notnull " json:"id"`                     // 主键ID
	RoleName  string    `xorm:"varchar(32) 'role_name' comment('角色名称') notnull " json:"role_name"`                  // 角色名称
	RoleKey   string    `xorm:"varchar(64) 'role_key' comment('角色关键字') notnull " json:"role_key"`                   // 角色关键字
	CreatedAt time.Time `xorm:"timestamp 'created_at' created  comment('创建时间') default NULL " json:"created_at"`    // 创建时间
	Deleted   int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"` // 是否删除 0 未删除 1 已删除
	UpdatedAt time.Time `xorm:"timestamp 'updated_at' updated comment('更新时间') default NULL " json:"updated_at"`     // 更新时间
	DeletedAt time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`      // 删除时间
}

func (s *ScaAuthRole) TableName() string {
	return "sca_auth_role"
}
