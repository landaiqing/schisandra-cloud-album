package middleware

import (
	"github.com/zeromicro/go-zero/rest/httpx"
	"net/http"
)

func UnsignedCallbackMiddleware() func(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
	return func(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
		httpx.WriteJsonCtx(r.Context(), w, http.StatusForbidden, "forbidden")
		return
	}
}
