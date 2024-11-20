package model

import "time"

type ScaFileFolder struct {
	Id             int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                                   // 主键
	FolderName     string    `xorm:"varchar(512) 'folder_name' comment('文件夹名称') default NULL " json:"folder_name"`                   // 文件夹名称
	ParentFolderId int64     `xorm:"bigint(20) 'parent_folder_id' comment('父文件夹编号') default NULL " json:"parent_folder_id"`          // 父文件夹编号
	FolderAddr     string    `xorm:"varchar(1024) 'folder_addr' comment('文件夹名称') default NULL " json:"folder_addr"`                  // 文件夹名称
	UserId         string    `xorm:"varchar(20) 'user_id' comment('用户编号') default NULL " json:"user_id"`                             // 用户编号
	FolderSource   int32     `xorm:"int(11) 'folder_source' comment('文件夹来源 0相册 1 评论') default NULL " json:"folder_source"`           // 文件夹来源 0相册 1 评论
	CreatedAt      time.Time `xorm:"datetime 'created_time' created comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdatedAt      time.Time `xorm:"datetime 'update_time' updated comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"`   // 更新时间
	Deleted        int32     `xorm:"int(11) 'deleted' comment('是否删除 0 未删除 1 已删除') default 0 " json:"deleted"`                        // 是否删除 0 未删除 1 已删除
	DeletedAt      time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`                  // 删除时间

}

func (s *ScaFileFolder) TableName() string {
	return "sca_file_folder"
}
