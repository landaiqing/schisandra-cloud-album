package model

import "time"

type ScaAuthUser struct {
	Id        int64     `xorm:"bigint(20) 'id' unique comment('自增ID') pk autoincr notnull " json:"id"`                        // 自增ID
	UID       string    `xorm:"varchar(20) 'uid' comment('唯一ID') notnull " json:"uid"`                                        // 唯一ID
	Username  string    `xorm:"varchar(32) 'username' comment('用户名') default NULL " json:"username"`                          // 用户名
	Nickname  string    `xorm:"varchar(32) 'nickname' comment('昵称') default NULL " json:"nickname"`                           // 昵称
	Email     string    `xorm:"varchar(32) 'email' comment('邮箱') default NULL " json:"email"`                                 // 邮箱
	Phone     string    `xorm:"varchar(32) 'phone' comment('电话') default NULL " json:"phone"`                                 // 电话
	Password  string    `xorm:"varchar(64) 'password' comment('密码') default NULL " json:"password"`                           // 密码
	Gender    int8      `xorm:"tinyint(4) 'gender' comment('性别') default NULL " json:"gender"`                                // 性别
	Avatar    string    `xorm:"longtext 'avatar' comment('头像') " json:"avatar"`                                               // 头像
	Status    int8      `xorm:"tinyint(4) 'status' comment('状态 0 正常 1 封禁') default 0 " json:"status"`                         // 状态 0 正常 1 封禁
	Introduce string    `xorm:"varchar(255) 'introduce' comment('介绍') default NULL " json:"introduce"`                        // 介绍
	CreatedAt time.Time `xorm:"timestamp  created 'created_at' comment('创建时间') default CURRENT_TIMESTAMP " json:"created_at"` // 创建时间
	Deleted   int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"`           // 是否删除 0 未删除 1 已删除
	Blog      string    `xorm:"varchar(30) 'blog' comment('博客') default NULL " json:"blog"`                                   // 博客
	Location  string    `xorm:"varchar(50) 'location' comment('地址') default NULL " json:"location"`                           // 地址
	Company   string    `xorm:"varchar(50) 'company' comment('公司') default NULL " json:"company"`                             // 公司
	UpdatedAt time.Time `xorm:"timestamp  updated 'updated_at' comment('更新时间') default NULL " json:"updated_at"`              // 更新时间
	DeletedAt time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`                // 删除时间

}

func (s *ScaAuthUser) TableName() string {
	return "sca_auth_user"
}
