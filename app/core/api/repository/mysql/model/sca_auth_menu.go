package model

import "time"

type ScaAuthMenu struct {
	Id        int64     `xorm:"bigint(20) 'id' comment('主键ID') pk autoincr notnull " json:"id"`                // 主键ID
	MenuName  string    `xorm:"varchar(64) 'menu_name' comment('名称') default NULL " json:"menu_name"`          // 名称
	ParentId  int64     `xorm:"bigint(20) 'parent_id' comment('父ID') default NULL " json:"parent_id"`          // 父ID
	Type      int8      `xorm:"tinyint(4) 'type' comment('类型 ') default 0 " json:"type"`                       // 类型
	Path      string    `xorm:"varchar(30) 'path' comment('路径') default NULL " json:"path"`                    // 路径
	Status    int8      `xorm:"tinyint(4) 'status' comment('状态 0 启用 1 停用') default 0 " json:"status"`          // 状态 0 启用 1 停用
	Icon      string    `xorm:"varchar(128) 'icon' comment('图标') default NULL " json:"icon"`                   // 图标
	MenuKey   string    `xorm:"varchar(64) 'menu_key' comment('关键字') default NULL " json:"menu_key"`           // 关键字
	Order     int32     `xorm:"int(11) 'order' comment('排序') default NULL " json:"order"`                      // 排序
	CreatedAt time.Time `xorm:"datetime 'created_at' comment('创建时间') default NULL " json:"created_at"`         // 创建时间
	UpdatedAt time.Time `xorm:"datetime 'updated_at' comment('更新时间') default NULL " json:"updated_at"`         // 更新时间
	Deleted   int32     `xorm:"int(11) 'deleted' comment('是否删除') default 0 " json:"deleted"`                   // 是否删除
	Remark    string    `xorm:"varchar(255) 'remark' comment('备注 描述') default NULL " json:"remark"`            // 备注 描述
	DeletedAt time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"` // 删除时间
}

func (s *ScaAuthMenu) TableName() string {
	return "sca_auth_menu"
}
