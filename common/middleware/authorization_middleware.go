package middleware

import (
	"github.com/zeromicro/go-zero/core/logx"
	"net/http"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
	"strconv"
	"time"
)

func AuthorizationMiddleware(w http.ResponseWriter, r *http.Request) {
	expireAtStr := r.Header.Get("X-Expire-At")
	if expireAtStr == "" {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusForbidden, "unauthorized"))
		return
	}
	expireAtInt, err := strconv.ParseInt(expireAtStr, 10, 64)
	if err != nil {
		logx.Errorf("Failed to parse X-Expire-At: %v", err)
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusForbidden, "unauthorized"))
		return
	}
	expireAt := time.Unix(expireAtInt, 0)
	currentTime := time.Now()

	remainingTime := expireAt.Sub(currentTime)
	if remainingTime < time.Minute*5 {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusUnauthorized, "token expired"))
		return
	}
}
