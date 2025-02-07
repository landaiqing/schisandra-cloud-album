package storage

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAlbumDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAlbumDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAlbumDetailLogic {
	return &GetAlbumDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAlbumDetailLogic) GetAlbumDetail(req *types.AlbumDetailListRequest) (resp string, err error) {
	// todo: add your logic here and delete this line

	return
}
