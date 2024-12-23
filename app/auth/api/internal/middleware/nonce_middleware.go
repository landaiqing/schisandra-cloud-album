package middleware

import (
	"github.com/redis/go-redis/v9"
	"net/http"
	"schisandra-album-cloud-microservices/common/middleware"
)

type NonceMiddleware struct {
	RedisClient *redis.Client
}

func NewNonceMiddleware(redisClient *redis.Client) *NonceMiddleware {
	return &NonceMiddleware{
		RedisClient: redisClient,
	}
}

func (m *NonceMiddleware) Handle(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		middleware.NonceMiddleware(w, r, m.RedisClient)
		next(w, r)
	}
}
