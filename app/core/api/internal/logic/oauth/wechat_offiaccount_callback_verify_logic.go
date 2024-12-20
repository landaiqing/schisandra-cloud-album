package oauth

import (
	"context"
	"net/http"

	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
)

type WechatOffiaccountCallbackVerifyLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWechatOffiaccountCallbackVerifyLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WechatOffiaccountCallbackVerifyLogic {
	return &WechatOffiaccountCallbackVerifyLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WechatOffiaccountCallbackVerifyLogic) WechatOffiaccountCallbackVerify(r *http.Request) (*http.Response, error) {
	rs, err := l.svcCtx.WechatOfficial.Server.VerifyURL(r)
	if err != nil {
		return nil, err
	}
	return rs, nil
}
