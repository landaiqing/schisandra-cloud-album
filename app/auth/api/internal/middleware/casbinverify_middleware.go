package middleware

import (
	"github.com/casbin/casbin/v2"
	"net/http"
	"schisandra-album-cloud-microservices/common/middleware"
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
		middleware.CasbinMiddleware(w, r, m.casbin)
		next(w, r)
	}
}
