package model

import "time"

type ScaAuthUserDevice struct {
	Id              int64     `xorm:"bigint(20) 'id' comment('主键ID') pk autoincr notnull " json:"id"`                     // 主键ID
	UserId          string    `xorm:"varchar(20) 'user_id' comment('用户ID') notnull " json:"user_id"`                      // 用户ID
	Ip              string    `xorm:"varchar(20) 'ip' comment('登录IP') notnull " json:"ip"`                                // 登录IP
	Location        string    `xorm:"varchar(20) 'location' comment('地址') notnull " json:"location"`                      // 地址
	Agent           string    `xorm:"varchar(255) 'agent' comment('设备信息') notnull " json:"agent"`                         // 设备信息
	CreatedAt       time.Time `xorm:"timestamp 'created_at' created comment('创建时间') default NULL " json:"created_at"`     // 创建时间
	Deleted         int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"` // 是否删除 0 未删除 1 已删除
	Browser         string    `xorm:"varchar(20) 'browser' comment('浏览器') notnull " json:"browser"`                       // 浏览器
	OperatingSystem string    `xorm:"varchar(20) 'operating_system' comment('操作系统') notnull " json:"operating_system"`    // 操作系统
	BrowserVersion  string    `xorm:"varchar(20) 'browser_version' comment('浏览器版本') notnull " json:"browser_version"`     // 浏览器版本
	Mobile          bool      `xorm:"tinyint(1) 'mobile' comment('是否为手机 0否1是') notnull " json:"mobile"`                   // 是否为手机 0否1是
	Bot             bool      `xorm:"tinyint(1) 'bot' comment('是否为bot 0否1是') notnull " json:"bot"`                        // 是否为bot 0否1是
	Mozilla         string    `xorm:"varchar(10) 'mozilla' comment('火狐版本') notnull " json:"mozilla"`                      // 火狐版本
	Platform        string    `xorm:"varchar(20) 'platform' comment('平台') notnull " json:"platform"`                      // 平台
	EngineName      string    `xorm:"varchar(20) 'engine_name' comment('引擎名称') notnull " json:"engine_name"`              // 引擎名称
	EngineVersion   string    `xorm:"varchar(20) 'engine_version' comment('引擎版本') notnull " json:"engine_version"`        // 引擎版本
	UpdatedAt       time.Time `xorm:"timestamp 'updated_at' updated comment('更新时间') default NULL " json:"updated_at"`     // 更新时间
	DeletedAt       time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`      // 删除时间

}

func (s *ScaAuthUserDevice) TableName() string {
	return "sca_auth_user_device"
}
