package oauth

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
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

func (l *GetQqOauthUrlLogic) GetQqOauthUrl(req *types.OAuthRequest) (resp *types.Response, err error) {
	clientId := l.svcCtx.Config.OAuth.QQ.ClientID
	redirectURI := l.svcCtx.Config.OAuth.QQ.RedirectURI
	url := "https://graph.qq.com/oauth2.0/authorize?response_type=code&client_id=" + clientId + "&redirect_uri=" + redirectURI + "&state=" + req.State
	return response.SuccessWithData(url), nil
}
