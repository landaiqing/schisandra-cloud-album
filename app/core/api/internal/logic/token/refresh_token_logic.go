package token

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
type AccessToken struct {
	AccessToken string `json:"access_token"`
	ExpireAt    int64  `json:"expire_at"`
}

func NewRefreshTokenLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RefreshTokenLogic {
	return &RefreshTokenLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *RefreshTokenLogic) RefreshToken(r *http.Request) (resp *types.Response, err error) {
	userId := r.Header.Get(constant.UID_HEADER_KEY)
	if userId == "" {
		return response.ErrorWithCode(403), nil
	}
	// 从redis中获取refresh token
	tokenData := l.svcCtx.RedisClient.Get(l.ctx, constant.UserTokenPrefix+userId).Val()
	if tokenData == "" {
		return response.ErrorWithCode(403), nil
	}
	redisTokenData := types.RedisToken{}
	err = json.Unmarshal([]byte(tokenData), &redisTokenData)
	if err != nil {
		return nil, err
	}
	// 判断是否已经被吊销
	if redisTokenData.Revoked {
		return response.ErrorWithCode(403), nil
	}
	// 判断是否是同一个设备
	if redisTokenData.AllowAgent != r.UserAgent() {
		return response.ErrorWithCode(403), nil
	}
	// 判断refresh token是否在有效期内
	refreshToken, result := jwt.ParseRefreshToken(l.svcCtx.Config.Auth.AccessSecret, redisTokenData.RefreshToken)
	if !result {
		return response.ErrorWithCode(403), nil
	}
	// 生成新的access token
	accessToken, expireAt := jwt.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
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
	token := AccessToken{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
	}
	return response.SuccessWithData(token), nil
}
