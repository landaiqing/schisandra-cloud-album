package middleware

import (
	"github.com/casbin/casbin/v2"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/xhttp"
)

type CasbinVerifyMiddleware struct {
	casbin *casbin.SyncedCachedEnforcer
}

func NewCasbinVerifyMiddleware(casbin *casbin.SyncedCachedEnforcer) *CasbinVerifyMiddleware {
	return &CasbinVerifyMiddleware{
		casbin: casbin,
	}
}

func (m *CasbinVerifyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Header.Get(constant.UID_HEADER_KEY)
		correct, err := m.casbin.Enforce(userId, r.URL.Path, r.Method)
		if err != nil || !correct {
			xhttp.JsonBaseResponseCtx(r.Context(), w, errors.New(http.StatusNotFound, "not found"))
			return
		}
		next(w, r)
	}
}
