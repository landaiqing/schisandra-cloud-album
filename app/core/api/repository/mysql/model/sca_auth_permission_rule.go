package model

type ScaAuthPermissionRule struct {
	Id    int32  `xorm:"int(11) 'id' pk autoincr notnull " json:"id"`
	Ptype string `xorm:"varchar(100) 'ptype' default NULL " json:"ptype"`
	V0    string `xorm:"varchar(100) 'v0' default NULL " json:"v0"`
	V1    string `xorm:"varchar(100) 'v1' default NULL " json:"v1"`
	V2    string `xorm:"varchar(100) 'v2' default NULL " json:"v2"`
	V3    string `xorm:"varchar(100) 'v3' default NULL " json:"v3"`
	V4    string `xorm:"varchar(100) 'v4' default NULL " json:"v4"`
	V5    string `xorm:"varchar(100) 'v5' default NULL " json:"v5"`
}

func (s *ScaAuthPermissionRule) TableName() string {
	return "sca_auth_permission_rule"
}
