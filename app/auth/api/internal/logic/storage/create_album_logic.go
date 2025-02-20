package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewCreateAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateAlbumLogic {
	return &CreateAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *CreateAlbumLogic) CreateAlbum(req *types.AlbumCreateRequest) (resp *types.AlbumCreateResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageAlbum := &model.ScaStorageAlbum{
		UserID:    uid,
		AlbumName: req.Name,
		AlbumType: constant.AlbumTypeMine,
	}
	err = l.svcCtx.DB.ScaStorageAlbum.Create(storageAlbum)
	if err != nil {
		return nil, err
	}
	return &types.AlbumCreateResponse{ID: storageAlbum.ID}, nil
}
