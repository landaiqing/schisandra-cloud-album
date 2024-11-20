package model

import "time"

type ScaFileType struct {
	Id        int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                                 // 主键
	TypeName  string    `xorm:"varchar(100) 'type_name' comment('类型名称') default NULL " json:"type_name"`                      // 类型名称
	MimeType  string    `xorm:"varchar(50) 'mime_type' comment('MIME 类型') default NULL " json:"mime_type"`                    // MIME 类型
	Status    int32     `xorm:"int(11) 'status' comment('类型状态') default NULL " json:"status"`                                 // 类型状态
	CreatedAt time.Time `xorm:"datetime 'created_at' created comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdatedAt time.Time `xorm:"datetime 'updated_at' updated  comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"` // 更新时间
	Deleted   int32     `xorm:"int(11) 'deleted' comment('是否删除 0 未删除 1 已删除') default 0 " json:"deleted"`                      // 是否删除 0 未删除 1 已删除
	DeletedAt time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`                // 删除时间

}

func (s *ScaFileType) TableName() string {
	return "sca_file_type"
}
