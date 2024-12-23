package oauth

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
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

func (l *GetGithubOauthUrlLogic) GetGithubOauthUrl(req *types.OAuthRequest) (resp string, err error) {
	clientId := l.svcCtx.Config.OAuth.Github.ClientID
	redirectUrl := l.svcCtx.Config.OAuth.Github.RedirectURI
	url := "https://github.com/login/oauth/authorize?client_id=" + clientId + "&redirect_uri=" + redirectUrl + "&state=" + req.State
	return url, nil
}
