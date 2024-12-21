package middleware

import (
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"strconv"
	"time"
)

type AuthorizationMiddleware struct {
}

func NewAuthorizationMiddleware() *AuthorizationMiddleware {
	return &AuthorizationMiddleware{}
}

func (m *AuthorizationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		expireAtStr := r.Header.Get("X-Expire-At")
		if expireAtStr == "" {
			httpx.OkJson(w, response.ErrorWithCodeMessage(http.StatusForbidden, "unauthorized"))
			return
		}
		expireAtInt, err := strconv.ParseInt(expireAtStr, 10, 64)
		if err != nil {
			logx.Errorf("Failed to parse X-Expire-At: %v", err)
			httpx.OkJson(w, response.ErrorWithCodeMessage(http.StatusForbidden, "unauthorized"))
			return
		}
		expireAt := time.Unix(expireAtInt, 0)
		currentTime := time.Now()

		remainingTime := expireAt.Sub(currentTime)
		if remainingTime < time.Minute*5 {
			httpx.OkJson(w, response.ErrorWithCodeMessage(http.StatusUnauthorized, "token about to expire, refresh"))
			return
		}
		next(w, r)
	}
}
