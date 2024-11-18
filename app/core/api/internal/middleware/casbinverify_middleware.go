package middleware

import (
	"net/http"

	"github.com/casbin/casbin/v2"
	"github.com/rbcervilla/redisstore/v9"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

type CasbinVerifyMiddleware struct {
	casbin  *casbin.CachedEnforcer
	session *redisstore.RedisStore
}

func NewCasbinVerifyMiddleware(casbin *casbin.CachedEnforcer, session *redisstore.RedisStore) *CasbinVerifyMiddleware {
	return &CasbinVerifyMiddleware{
		casbin:  casbin,
		session: session,
	}
}

func (m *CasbinVerifyMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := m.session.Get(r, constant.SESSION_KEY)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		userId, ok := session.Values["uid"].(string)
		if !ok {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		correct, err := m.casbin.Enforce(userId, r.URL.Path, r.Method)
		if err != nil || !correct {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next(w, r)
	}
}
