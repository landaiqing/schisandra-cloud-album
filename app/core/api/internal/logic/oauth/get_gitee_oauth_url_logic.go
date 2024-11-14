package oauth

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGiteeOauthUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGiteeOauthUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGiteeOauthUrlLogic {
	return &GetGiteeOauthUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGiteeOauthUrlLogic) GetGiteeOauthUrl() (resp *types.Response, err error) {
	clientID := l.svcCtx.Config.OAuth.Gitee.ClientID
	redirectURI := l.svcCtx.Config.OAuth.Gitee.RedirectURI
	url := "https://gitee.com/oauth/authorize?client_id=" + clientID + "&redirect_uri=" + redirectURI + "&response_type=code"
	return response.SuccessWithData(url), nil
}
