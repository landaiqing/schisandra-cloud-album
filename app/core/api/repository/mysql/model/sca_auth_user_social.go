package model

import "time"

type ScaAuthUserSocial struct {
	Id        int64     `xorm:"bigint(20) 'id' comment('主键ID') pk autoincr notnull " json:"id"`                     // 主键ID
	UserId    string    `xorm:"varchar(20) 'user_id' comment('用户ID') notnull " json:"user_id"`                      // 用户ID
	OpenId    string    `xorm:"varchar(50) 'open_id' comment('第三方用户的 open id') notnull " json:"open_id"`            // 第三方用户的 open id
	Source    string    `xorm:"varchar(10) 'source' comment('第三方用户来源') notnull " json:"source"`                     // 第三方用户来源
	Status    int64     `xorm:"bigint(20) 'status' comment('状态 0正常 1 封禁') notnull default 0 " json:"status"`        // 状态 0正常 1 封禁
	CreatedAt time.Time `xorm:"timestamp 'created_at' created comment('创建时间') default NULL " json:"created_at"`     // 创建时间
	Deleted   int8      `xorm:"tinyint(4) 'deleted' comment('是否删除 0 未删除 1 已删除') notnull default 0 " json:"deleted"` // 是否删除 0 未删除 1 已删除
	UpdatedAt time.Time `xorm:"timestamp 'updated_at'updated  comment('更新时间') default NULL " json:"updated_at"`     // 更新时间
	DeletedAt time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`      // 删除时间

}

func (s *ScaAuthUserSocial) TableName() string {
	return "sca_auth_user_social"
}
