package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
)

type CasbinVerifyMiddleware struct {
	casbin *casbin.CachedEnforcer
}

func NewCasbinVerifyMiddleware(casbin *casbin.CachedEnforcer) *CasbinVerifyMiddleware {
	return &CasbinVerifyMiddleware{
		casbin: casbin,
	}
}

func (m *CasbinVerifyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("user_id")
		correct, err := m.casbin.Enforce(userId, r.URL.Path, r.Method)
		if err != nil || !correct {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
