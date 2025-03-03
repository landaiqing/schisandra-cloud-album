package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AddImageToAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAddImageToAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AddImageToAlbumLogic {
	return &AddImageToAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AddImageToAlbumLogic) AddImageToAlbum(req *types.AddImageToAlbumRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	update, err := storageInfo.Where(storageInfo.UserID.Eq(uid),
		storageInfo.ID.In(req.IDS...),
		storageInfo.Provider.Eq(req.Provider),
		storageInfo.Bucket.Eq(req.Bucket)).Update(storageInfo.AlbumID, req.AlbumID)
	if err != nil {
		return "", err
	}
	if update.RowsAffected == 0 {
		return "", errors.New("no image found")
	}
	return "success", nil
}
