package redis_session

import (
	"context"
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logc"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/repository/redis_session/types"
)

func NewRedisSession(addr string, password string) *redisstore.RedisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})
	store, err := redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		logc.Error(context.Background(), err)
	}
	store.KeyPrefix(constant.UserSessionPrefix)
	store.Options(sessions.Options{
		Path: "/",
		// Domain: global.CONFIG.System.Web,
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	})
	gob.Register(types.SessionData{})
	return store
}
