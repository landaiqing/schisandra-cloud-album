package system

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllStorageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllStorageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllStorageListLogic {
	return &GetAllStorageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllStorageListLogic) GetAllStorageList() (resp *types.AllStorageListResponse, err error) {
	// todo: add your logic here and delete this line

	return
}
