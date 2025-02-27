package storage

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadAlbumLogic {
	return &DownloadAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadAlbumLogic) DownloadAlbum(req *types.DownloadAlbumRequest) (resp string, err error) {
	// todo: download album logic

	return
}
