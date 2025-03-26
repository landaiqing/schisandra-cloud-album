package system

import (
	"context"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserThirdPartyLoginListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserThirdPartyLoginListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserThirdPartyLoginListLogic {
	return &GetUserThirdPartyLoginListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserThirdPartyLoginListLogic) GetUserThirdPartyLoginList() (resp *types.UserThirdPartyLoginListResponse, err error) {

	userSocial := l.svcCtx.DB.ScaAuthUserSocial
	var userSocialList []types.UserThirdPartyLoginMeta
	err = userSocial.Scan(&userSocialList)
	if err != nil {
	}
	return &types.UserThirdPartyLoginListResponse{Records: userSocialList}, nil
}
