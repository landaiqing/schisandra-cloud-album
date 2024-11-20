package model

import "time"

type ScaFileInfo struct {
	Id         int64     `xorm:"bigint(20) 'id' comment('主键') pk autoincr notnull " json:"id"`                                 // 主键
	FileName   string    `xorm:"varchar(50) 'file_name' comment('文件名') default NULL " json:"file_name"`                        // 文件名
	FileSize   float64   `xorm:"double 'file_size' comment('文件大小') default NULL " json:"file_size"`                            // 文件大小
	FileTypeId int64     `xorm:"bigint(20) 'file_type_id' comment('文件类型编号') default NULL " json:"file_type_id"`                // 文件类型编号
	UploadTime time.Time `xorm:"datetime 'upload_time' comment('上传时间') default NULL " json:"upload_time"`                      // 上传时间
	FolderId   int64     `xorm:"bigint(20) 'folder_id' comment('文件夹编号') default NULL " json:"folder_id"`                       // 文件夹编号
	UserId     string    `xorm:"varchar(20) 'user_id' comment('用户编号') default NULL " json:"user_id"`                           // 用户编号
	FileSource int32     `xorm:"int(11) 'file_source' comment('文件来源 0 相册 1 评论') default NULL " json:"file_source"`             // 文件来源 0 相册 1 评论
	Status     int32     `xorm:"int(11) 'status' comment('文件状态') default NULL " json:"status"`                                 // 文件状态
	CreatedAt  time.Time `xorm:"datetime 'created_at' created comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdatedAt  time.Time `xorm:"datetime 'updated_at' updated comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"`  // 更新时间
	Deleted    int32     `xorm:"int(11) 'deleted' comment('是否删除 0 未删除 1 已删除') default 0 " json:"deleted"`                      // 是否删除 0 未删除 1 已删除
	DeletedAt  time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"`                // 删除时间

}

func (s *ScaFileInfo) TableName() string {
	return "sca_file_info"
}
