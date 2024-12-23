package oauth

import (
	"context"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
)

type GetQqOauthUrlLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetQqOauthUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetQqOauthUrlLogic {
	return &GetQqOauthUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetQqOauthUrlLogic) GetQqOauthUrl(req *types.OAuthRequest) (resp string, err error) {
	clientId := l.svcCtx.Config.OAuth.QQ.ClientID
	redirectURI := l.svcCtx.Config.OAuth.QQ.RedirectURI
	url := "https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=" + clientId + "&redirect_uri=" + redirectURI + "&state=" + req.State
	return url, nil
}
