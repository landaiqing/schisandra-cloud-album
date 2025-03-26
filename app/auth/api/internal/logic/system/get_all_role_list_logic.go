package system

import (
	"context"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllRoleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllRoleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllRoleListLogic {
	return &GetAllRoleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllRoleListLogic) GetAllRoleList() (resp *types.RoleListResponse, err error) {
	authRole := l.svcCtx.DB.ScaAuthRole
	var roles []types.RoleMeta
	err = authRole.Scan(&roles)
	if err != nil {
		return nil, err
	}
	return &types.RoleListResponse{Records: roles}, nil
}
