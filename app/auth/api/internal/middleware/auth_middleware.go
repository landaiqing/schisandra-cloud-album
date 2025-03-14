package middleware

import (
	"context"
	"github.com/redis/go-redis/v9"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
)

type AuthMiddleware struct {
	RedisClient *redis.Client
	ctx         context.Context
}

func NewAuthMiddleware(redisClient *redis.Client) *AuthMiddleware {
	return &AuthMiddleware{
		RedisClient: redisClient,
		ctx:         context.Background(),
	}
}

func (m *AuthMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get(constant.UID_HEADER_KEY)
		if userId == "" {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusNotFound, "user id not found in header"))
			return
		}
		cacheToken := constant.UserTokenPrefix + userId
		result, err := m.RedisClient.Get(m.ctx, cacheToken).Result()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusInternalServerError, err.Error()))
			return
		}
		if result == "" {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusForbidden, "access denied"))
			return
		}
		next(w, r)
	}
}
