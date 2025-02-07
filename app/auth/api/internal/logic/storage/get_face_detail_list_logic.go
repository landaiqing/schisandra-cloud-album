package storage

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetFaceDetailListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetFaceDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetFaceDetailListLogic {
	return &GetFaceDetailListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetFaceDetailListLogic) GetFaceDetailList(req *types.FaceDetailListRequest) (resp string, err error) {
	// todo: add your logic here and delete this line

	return
}
