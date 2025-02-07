package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RenameAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRenameAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RenameAlbumLogic {
	return &RenameAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RenameAlbumLogic) RenameAlbum(req *types.AlbumRenameRequest) (resp *types.AlbumRenameResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageAlbum := l.svcCtx.DB.ScaStorageAlbum
	info, err := storageAlbum.Where(storageAlbum.ID.Eq(req.ID), storageAlbum.UserID.Eq(uid)).Update(storageAlbum.AlbumName, req.Name)
	if err != nil {
		return nil, err
	}
	if info.RowsAffected == 0 {
		return nil, errors.New("album not found")
	}
	return &types.AlbumRenameResponse{
		ID:   req.ID,
		Name: req.Name,
	}, nil
}
