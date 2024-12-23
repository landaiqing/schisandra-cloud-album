package middleware

import (
	"net/http"
	"schisandra-album-cloud-microservices/common/middleware"
)

type AuthorizationMiddleware struct {
}

func NewAuthorizationMiddleware() *AuthorizationMiddleware {
	return &AuthorizationMiddleware{}
}

func (m *AuthorizationMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.AuthorizationMiddleware(w, r)
		next(w, r)
	}
}
