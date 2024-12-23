package middleware

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func UnauthorizedCallbackMiddleware() func(w http.ResponseWriter, r *http.Request, err error) {
	return func(w http.ResponseWriter, r *http.Request, err error) {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusUnauthorized, err.Error()))
		return
	}
}
