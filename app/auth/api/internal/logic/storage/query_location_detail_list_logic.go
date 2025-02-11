package storage

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryLocationDetailListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLocationDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLocationDetailListLogic {
	return &QueryLocationDetailListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryLocationDetailListLogic) QueryLocationDetailList(req *types.LocationDetailListRequest) (resp *types.LocationDetailListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
