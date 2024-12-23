package token

import (
	"context"
	"encoding/json"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	jwt2 "schisandra-album-cloud-microservices/common/jwt"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

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

func (l *RefreshTokenLogic) RefreshToken(r *http.Request) (resp *types.RefreshTokenResponse, err error) {
	userId := r.Header.Get(constant.UID_HEADER_KEY)
	if userId == "" {
		return nil, errors.New(http.StatusForbidden, "user id is empty")
	}
	// 从redis中获取refresh token
	tokenData := l.svcCtx.RedisClient.Get(l.ctx, constant.UserTokenPrefix+userId).Val()
	if tokenData == "" {
		return nil, errors.New(http.StatusForbidden, "refresh token is empty")
	}
	redisTokenData := types.RedisToken{}
	err = json.Unmarshal([]byte(tokenData), &redisTokenData)
	if err != nil {
		return nil, err
	}
	// 判断是否已经被吊销
	if redisTokenData.Revoked {
		return nil, errors.New(http.StatusForbidden, "refresh token is revoked")
	}
	// 判断是否是同一个设备
	if redisTokenData.AllowAgent != r.UserAgent() {
		return nil, errors.New(http.StatusForbidden, "refresh token is not allowed for this agent")
	}
	// 判断refresh token是否在有效期内
	refreshToken, result := jwt2.ParseRefreshToken(l.svcCtx.Config.Auth.AccessSecret, redisTokenData.RefreshToken)
	if !result {
		return nil, errors.New(http.StatusForbidden, "refresh token is invalid")
	}
	// 生成新的access token
	accessToken, expireAt := jwt2.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, jwt2.AccessJWTPayload{
		UserID: refreshToken.UserID,
		Type:   constant.JWT_TYPE_ACCESS,
	})
	// 更新redis中的access token
	redisToken := types.RedisToken{
		AccessToken:  accessToken,
		RefreshToken: redisTokenData.RefreshToken,
		UID:          refreshToken.UserID,
		Revoked:      false,
		GeneratedAt:  redisTokenData.GeneratedAt,
		AllowAgent:   redisTokenData.AllowAgent,
		GeneratedIP:  redisTokenData.GeneratedIP,
		UpdatedAt:    time.Now().Format(constant.TimeFormat),
	}
	err = l.svcCtx.RedisClient.Set(l.ctx, constant.UserTokenPrefix+refreshToken.UserID, redisToken, time.Hour*24*7).Err()
	if err != nil {
		return nil, err
	}
	token := &types.RefreshTokenResponse{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
	}
	return token, nil
}
