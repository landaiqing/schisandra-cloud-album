package middleware

import (
	"net/http"

	"schisandra-album-cloud-microservices/app/core/api/common/middleware"
)

type SecurityHeadersMiddleware struct {
}

func NewSecurityHeadersMiddleware() *SecurityHeadersMiddleware {
	return &SecurityHeadersMiddleware{}
}

func (m *SecurityHeadersMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.SecurityHeadersMiddleware(w, r)
		next(w, r)
	}
}
