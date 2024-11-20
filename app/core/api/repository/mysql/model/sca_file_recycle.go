package model

import "time"

type ScaFileRecycle struct {
	Id           int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                     // 主键
	FileId       int64     `xorm:"bigint(20) 'file_id' comment('文件编号') default NULL " json:"file_id"`                // 文件编号
	FolderId     int64     `xorm:"bigint(20) 'folder_id' comment('文件夹编号') default NULL " json:"folder_id"`           // 文件夹编号
	Type         int32     `xorm:"int(11) 'type' comment('类型 0 文件 1 文件夹') default NULL " json:"type"`                // 类型 0 文件 1 文件夹
	UserId       string    `xorm:"varchar(20) 'user_id' comment('用户编号') default NULL " json:"user_id"`               // 用户编号
	DeletedAt    time.Time `xorm:"datetime 'deleted_at' comment('删除时间') default NULL " json:"deleted_at"`            // 删除时间
	OriginalPath string    `xorm:"varchar(1024) 'original_path' comment('原始路径') default NULL " json:"original_path"` // 原始路径
	Deleted      int32     `xorm:"int(11) 'deleted' comment('是否被永久删除 0否 1是') default NULL " json:"deleted"`          // 是否被永久删除 0否 1是
	FileSource   int32     `xorm:"int(11) 'file_source' comment('文件来源 0 相册 1 评论') default NULL " json:"file_source"` // 文件来源 0 相册 1 评论

}

func (s *ScaFileRecycle) TableName() string {
	return "sca_file_recycle"
}
