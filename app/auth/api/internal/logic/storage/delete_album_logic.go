package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteAlbumLogic {
	return &DeleteAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteAlbumLogic) DeleteAlbum(req *types.AlbumDeleteRequest) (resp string, err error) {
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

	info, err := tx.ScaStorageAlbum.Where(tx.ScaStorageAlbum.ID.Eq(req.ID), tx.ScaStorageAlbum.UserID.Eq(uid)).Delete()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if info.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("album not found")
	}
	storageInfo := tx.ScaStorageInfo
	_, err = storageInfo.Where(storageInfo.AlbumID.Eq(req.ID), storageInfo.UserID.Eq(uid)).Delete()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	return "success", nil
}
