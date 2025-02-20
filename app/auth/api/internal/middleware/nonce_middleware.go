package middleware

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
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
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
			return
		}
		if len(nonce) != 32 {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
			return
		}
		result := m.RedisClient.Get(r.Context(), constant.SystemApiNoncePrefix+nonce).Val()
		if result != "" {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
			return
		}
		err := m.RedisClient.Set(r.Context(), constant.SystemApiNoncePrefix+nonce, nonce, time.Minute*1).Err()
		if err != nil {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusInternalServerError, "internal server error!"))
			return
		}
		next(w, r)
	}
}
