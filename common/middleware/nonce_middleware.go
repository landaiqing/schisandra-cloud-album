package middleware

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
	"time"
)

func NonceMiddleware(w http.ResponseWriter, r *http.Request, redisClient *redis.Client) {
	nonce := r.Header.Get("X-Nonce")
	if nonce == "" {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
		return
	}
	if len(nonce) != 32 {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
		return
	}
	result := redisClient.Get(r.Context(), constant.SystemApiNoncePrefix+nonce).Val()
	if result != "" {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusBadRequest, "bad request!"))
		return
	}
	err := redisClient.Set(r.Context(), constant.SystemApiNoncePrefix+nonce, nonce, time.Minute*1).Err()
	if err != nil {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusInternalServerError, "internal server error!"))
		return
	}
}
