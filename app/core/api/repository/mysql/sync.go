package mysql

import (
	"xorm.io/xorm"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
)

// SyncDatabase function is used to create all tables in the database if they don't exist.
func SyncDatabase(engine *xorm.Engine) error {
	err := engine.Sync2(
		new(model.ScaAuthUser),
		new(model.ScaCommentReply),
		new(model.ScaAuthUserSocial),
		new(model.ScaAuthUserDevice),
		new(model.ScaCommentLikes),
		new(model.ScaAuthRole),
		new(model.ScaFileRecycle),
		new(model.ScaFileType),
		new(model.ScaMessageReport),
		new(model.ScaUserFollows),
		new(model.ScaUserLevel),
		new(model.ScaUserMessage),
		new(model.ScaAuthMenu),
		new(model.ScaAuthPermissionRule),
		new(model.ScaFileFolder),
		new(model.ScaFileInfo),
	)
	if err != nil {
		return err
	}
	return nil
}
