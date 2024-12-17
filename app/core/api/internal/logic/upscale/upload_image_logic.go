package upscale

import (
	"context"
	"encoding/json"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/logic/websocket"

	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type UploadImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadImageLogic {
	return &UploadImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadImageLogic) UploadImage(r *http.Request, req *types.UploadRequest) (resp *types.Response, err error) {
	token, ok := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, req.AccessToken)
	if !ok {
		return response.ErrorWithI18n(l.ctx, "upload.uploadError"), nil
	}
	if token.UserID != req.UserId {
		return response.ErrorWithI18n(l.ctx, "upload.uploadError"), nil
	}

	correct, err := l.svcCtx.CasbinEnforcer.Enforce(req.UserId, r.URL.Path, r.Method)
	if err != nil || !correct {
		return response.ErrorWithI18n(l.ctx, "upload.uploadError"), err
	}

	data, err := json.Marshal(response.SuccessWithData(req.Image))
	if err != nil {
		return nil, err
	}
	err = websocket.FileWebSocketHandler.SendMessageToClient(req.UserId, data)
	if err != nil {
		return nil, err
	}
	return response.Success(), nil
}
