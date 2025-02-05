package middleware

import (
	"github.com/casbin/casbin/v2"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
)

func CasbinMiddleware(w http.ResponseWriter, r *http.Request, casbin *casbin.SyncedCachedEnforcer) {
	userId := r.Header.Get(constant.UID_HEADER_KEY)
	correct, err := casbin.Enforce(userId, r.URL.Path, r.Method)
	if err != nil || !correct {
		xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusNotFound, "not found"))
		return
	}
}
