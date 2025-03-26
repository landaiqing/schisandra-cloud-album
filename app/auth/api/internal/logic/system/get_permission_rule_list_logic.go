package system

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPermissionRuleListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPermissionRuleListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPermissionRuleListLogic {
	return &GetPermissionRuleListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPermissionRuleListLogic) GetPermissionRuleList() (resp *types.PermissionRuleListResponse, err error) {

	permissionRule := l.svcCtx.DB.ScaAuthPermissionRule
	var permissionRuleList []types.PermissionRuleMeta
	err = permissionRule.Scan(&permissionRuleList)
	if err != nil {
		return nil, err
	}

	return &types.PermissionRuleListResponse{Records: permissionRuleList}, nil
}
