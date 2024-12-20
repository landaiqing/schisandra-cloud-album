package middleware

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest/httpx"

	"schisandra-album-cloud-microservices/app/core/api/common/response"
)

func UnauthorizedCallbackMiddleware() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		httpx.OkJsonCtx(r.Context(), w, response.ErrorWithCodeMessage(http.StatusUnauthorized, err.Error()))
		return
	}
}
