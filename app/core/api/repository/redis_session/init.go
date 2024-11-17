package redis_session

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
)

func NewRedisSession(client *redis.Client) *redisstore.RedisStore {
	store, err := redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		panic(err)
	}
	store.KeyPrefix(constant.UserSessionPrefix)
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	gob.Register(types.SessionData{})
	return store
}
