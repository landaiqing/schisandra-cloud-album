package user

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type RefreshTokenLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(r *http.Request) (resp *types.Response, err error) {
	session, err := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		return response.ErrorWithCode(403), err
	}
	sessionData, ok := session.Values[constant.SESSION_KEY]
	if !ok {
		return response.ErrorWithCode(403), err
	}
	data := types.SessionData{}
	err = json.Unmarshal(sessionData.([]byte), &data)
	if err != nil {
		return response.ErrorWithCode(403), err
	}
	refreshToken, result := jwt.ParseRefreshToken(l.svcCtx.Config.Auth.AccessSecret, data.RefreshToken)
	if !result {
		return response.ErrorWithCode(403), err
	}
	accessToken := jwt.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: refreshToken.UserID,
	})
	if accessToken == "" {
		return response.ErrorWithCode(403), err
	}
	redisToken := types.RedisToken{
		AccessToken: accessToken,
		UID:         refreshToken.UserID,
	}
	err = l.svcCtx.RedisClient.Set(l.ctx, constant.UserTokenPrefix+refreshToken.UserID, redisToken, time.Hour*24*7).Err()
	if err != nil {
		return response.ErrorWithCode(403), err
	}

	return response.SuccessWithData(accessToken), nil
}
