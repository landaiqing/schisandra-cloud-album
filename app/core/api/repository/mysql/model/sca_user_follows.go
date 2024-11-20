package model

import "time"

type ScaUserFollows struct {
	FollowerId string    `xorm:"varchar(255) 'follower_id' comment('关注者') notnull " json:"follower_id"`                        // 关注者
	FolloweeId string    `xorm:"varchar(255) 'followee_id' comment('被关注者') notnull " json:"followee_id"`                       // 被关注者
	Status     uint8     `xorm:"tinyint(3) UNSIGNED 'status' comment('关注状态（0 未互关 1 互关）') notnull default 0 " json:"status"`    // 关注状态（0 未互关 1 互关）
	CreatedAt  time.Time `xorm:"datetime 'created_at' created comment('创建时间') default CURRENT_TIMESTAMP " json:"created_time"` // 创建时间
	UpdatedAt  time.Time `xorm:"datetime 'updated_at' updated comment('更新时间') default CURRENT_TIMESTAMP " json:"update_time"`  // 更新时间
	Id         int64     `xorm:"bigint(20) 'id' pk autoincr notnull " json:"id"`
	DeletedAt  time.Time `xorm:"datetime 'deleted_at' deleted comment('删除时间') default NULL " json:"deleted_at"` // 删除时间

}

func (s *ScaUserFollows) TableName() string {
	return "sca_user_follows"
}
