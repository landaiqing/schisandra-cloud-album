package middleware

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
)

type AuthorizationMiddleware struct {
	Redis *redis.Client
}

func NewAuthorizationMiddleware(redis *redis.Client) *AuthorizationMiddleware {
	return &AuthorizationMiddleware{
		Redis: redis,
	}
}

func (m *AuthorizationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id").(string)
		redisToken := m.Redis.Get(r.Context(), constant.UserTokenPrefix+userId).Val()
		if redisToken == "" {
			httpx.OkJson(w, response.ErrorWithCodeMessage(403, "unauthorized"))
			return
		}
		next(w, r)
	}
}
