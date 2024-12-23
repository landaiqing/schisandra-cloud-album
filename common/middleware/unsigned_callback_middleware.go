package middleware

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func UnsignedCallbackMiddleware() func(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
	return func(w http.ResponseWriter, r *http.Request, next http.Handler, strict bool, code int) {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusForbidden, "Unsigned callback not allowed"))
		return
	}
}
