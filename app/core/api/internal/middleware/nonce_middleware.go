package middleware

import (
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"time"
)

type NonceMiddleware struct {
	RedisClient *redis.Client
}

func NewNonceMiddleware(redisClient *redis.Client) *NonceMiddleware {
	return &NonceMiddleware{
		RedisClient: redisClient,
	}
}

func (m *NonceMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		nonce := r.Header.Get("X-Nonce")
		if nonce == "" {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, "bad request!")
			return
		}
		if len(nonce) != 32 {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, "bad request!")
			return
		}
		result := m.RedisClient.Get(r.Context(), constant.SystemApiNoncePrefix+nonce).Val()
		if result != "" {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusBadRequest, "bad request!")
			return
		}
		err := m.RedisClient.Set(r.Context(), constant.SystemApiNoncePrefix+nonce, nonce, time.Minute*1).Err()
		if err != nil {
			httpx.WriteJsonCtx(r.Context(), w, http.StatusInternalServerError, "internal server error!")
			return
		}
		next(w, r)
	}
}
