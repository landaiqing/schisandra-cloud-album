package system

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserLoginLogListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserLoginLogListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserLoginLogListLogic {
	return &GetUserLoginLogListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserLoginLogListLogic) GetUserLoginLogList() (resp *types.UserLoginLogListResponse, err error) {
	userDevice := l.svcCtx.DB.ScaAuthUserDevice
	var userLoginLogs []types.UserLoginLogMeta
	err = userDevice.Scan(&userLoginLogs)
	if err != nil {
		return nil, err
	}

	return &types.UserLoginLogListResponse{Records: userLoginLogs}, nil
}
