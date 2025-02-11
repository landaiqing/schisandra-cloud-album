package storage

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryThingDetailListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryThingDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryThingDetailListLogic {
	return &QueryThingDetailListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryThingDetailListLogic) QueryThingDetailList(req *types.ThingDetailListRequest) (resp *types.ThingDetailListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
