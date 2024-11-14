package oauth

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetGithubOauthUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetGithubOauthUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetGithubOauthUrlLogic {
	return &GetGithubOauthUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetGithubOauthUrlLogic) GetGithubOauthUrl(req *types.OAuthRequest) (resp *types.Response, err error) {
	clientId := l.svcCtx.Config.OAuth.Github.ClientID
	redirectUrl := l.svcCtx.Config.OAuth.Github.RedirectURI
	url := "https://github.com/login/oauth/authorize?client_id=" + clientId + "&redirect_uri=" + redirectUrl + "&state=" + req.State
	return response.SuccessWithData(url), nil
}
