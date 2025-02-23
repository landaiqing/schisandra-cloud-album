package phone

import (
	"context"
	"encoding/json"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/logic/websocket"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/jwt"
	"schisandra-album-cloud-microservices/common/xhttp"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

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

func (l *UploadImageLogic) UploadImage(r *http.Request, req *types.UploadRequest) error {
	token, ok := jwt.ParseAccessToken(l.svcCtx.Config.Auth.AccessSecret, req.AccessToken)
	if !ok {
		return errors.New(http.StatusForbidden, "invalid access token")
	}
	if token.UserID != req.UserId {
		return errors.New(http.StatusForbidden, "invalid user id")
	}

	correct, err := l.svcCtx.CasbinEnforcer.Enforce(req.UserId, r.URL.Path, r.Method)
	if err != nil || !correct {
		return errors.New(http.StatusForbidden, "permission denied")
	}

	data, err := json.Marshal(xhttp.BaseResponse[string]{
		Data: req.Image,
		Msg:  "success",
		Code: http.StatusOK,
	})
	if err != nil {
		return errors.New(http.StatusForbidden, err.Error())
	}
	err = websocket.FileWebSocketHandler.SendMessageToClient(req.UserId, data)
	if err != nil {
		return errors.New(http.StatusForbidden, err.Error())
	}
	return nil
}
