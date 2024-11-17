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
		return nil, err
	}
	refreshSessionToken, ok := session.Values["refresh_token"].(string)
	if !ok {
		return response.ErrorWithCode(403), nil
	}
	userId, ok := session.Values["uid"].(string)
	if !ok {
		return response.ErrorWithCode(403), nil
	}
	tokenData := l.svcCtx.RedisClient.Get(l.ctx, constant.UserTokenPrefix+userId).Val()
	if tokenData == "" {
		return response.ErrorWithCode(403), nil
	}
	redisTokenData := types.RedisToken{}
	err = json.Unmarshal([]byte(tokenData), &redisTokenData)
	if err != nil {
		return nil, err
	}
	if redisTokenData.RefreshToken != refreshSessionToken {
		return response.ErrorWithCode(403), nil
	}

	refreshToken, result := jwt.ParseRefreshToken(l.svcCtx.Config.Auth.AccessSecret, refreshSessionToken)
	if !result {
		return response.ErrorWithCode(403), nil
	}
	accessToken := jwt.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: refreshToken.UserID,
	})
	if accessToken == "" {
		return response.ErrorWithCode(403), nil
	}
	redisToken := types.RedisToken{
		AccessToken:  accessToken,
		RefreshToken: refreshSessionToken,
		UID:          refreshToken.UserID,
	}
	err = l.svcCtx.RedisClient.Set(l.ctx, constant.UserTokenPrefix+refreshToken.UserID, redisToken, time.Hour*24*7).Err()
	if err != nil {
		return nil, err
	}

	return response.SuccessWithData(accessToken), nil
}
