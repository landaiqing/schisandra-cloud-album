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
	info, err := l.svcCtx.DB.ScaStorageAlbum.Where(l.svcCtx.DB.ScaStorageAlbum.ID.Eq(req.ID), l.svcCtx.DB.ScaStorageAlbum.UserID.Eq(uid)).Delete()
	if err != nil {
		return "", err
	}
	if info.RowsAffected == 0 {
		return "", errors.New("album not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	_, err = storageInfo.Where(storageInfo.AlbumID.Eq(req.ID), storageInfo.UserID.Eq(uid)).Update(storageInfo.AlbumID, 0)
	if err != nil {
		return "", err
	}
	return "success", nil
}
