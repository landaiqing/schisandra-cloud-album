package share

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/common/constant"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteShareRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteShareRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteShareRecordLogic {
	return &DeleteShareRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteShareRecordLogic) DeleteShareRecord(req *types.DeleteShareRecordRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}

	tx := l.svcCtx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	storageShare := tx.ScaStorageShare
	storageShareDeleted, err := storageShare.Where(storageShare.UserID.Eq(uid),
		storageShare.ID.Eq(req.ID),
		storageShare.InviteCode.Eq(req.InviteCode),
		storageShare.AlbumID.Eq(req.AlbumID)).
		Delete()
	if err != nil || storageShareDeleted.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("delete share record failed")
	}
	shareVisit := tx.ScaStorageShareVisit
	_, err = shareVisit.Where(shareVisit.ShareID.Eq(req.ID), shareVisit.UserID.Eq(uid)).Delete()
	if err != nil {
		tx.Rollback()
		return "", errors.New("delete share visit record failed")
	}
	storageAlbum := tx.ScaStorageAlbum
	albumDeleted, err := storageAlbum.Where(storageAlbum.ID.Eq(req.AlbumID), storageAlbum.UserID.Eq(uid)).Delete()
	if err != nil || albumDeleted.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("delete album record failed")
	}
	storageInfo := tx.ScaStorageInfo
	infoDeleted, err := storageInfo.Where(storageInfo.AlbumID.Eq(req.AlbumID), storageInfo.UserID.Eq(uid)).Delete()
	if err != nil || infoDeleted.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("delete storage info record failed")
	}
	// delete redis cache
	cacheKey := constant.ImageSharePrefix + req.InviteCode
	err = l.svcCtx.RedisClient.Del(l.ctx, cacheKey).Err()
	if err != nil {
		tx.Rollback()
		return "", errors.New("delete cache failed")
	}
	cacheVisitKey := constant.ImageShareVisitPrefix + req.InviteCode
	err = l.svcCtx.RedisClient.Del(l.ctx, cacheVisitKey).Err()
	if err != nil {
		tx.Rollback()
		return "", errors.New("delete cache visit failed")
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", errors.New("commit failed")
	}
	return
}
