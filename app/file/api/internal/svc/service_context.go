package svc

import (
	"github.com/casbin/casbin/v2"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	"schisandra-album-cloud-microservices/app/file/api/internal/config"
	"schisandra-album-cloud-microservices/app/file/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/file/api/model/mysql"
	"schisandra-album-cloud-microservices/app/file/api/model/mysql/query"
	"schisandra-album-cloud-microservices/common/casbinx"
	"schisandra-album-cloud-microservices/common/redisx"
)

type ServiceContext struct {
	Config                    config.Config
	NonceMiddleware           rest.Middleware
	SecurityHeadersMiddleware rest.Middleware
	DB                        *query.Query
	CasbinEnforcer            *casbin.SyncedCachedEnforcer
	RedisClient               *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redisx.NewRedis(c.Redis.Host, c.Redis.Pass, c.Redis.DB)
	db, queryDB := mysql.NewMySQL(c.Mysql.DataSource, c.Mysql.MaxOpenConn, c.Mysql.MaxIdleConn, redisClient)
	casbinEnforcer := casbinx.NewCasbin(db)
	return &ServiceContext{
		Config:                    c,
		NonceMiddleware:           middleware.NewNonceMiddleware(redisClient).Handle,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		DB:                        queryDB,
		CasbinEnforcer:            casbinEnforcer,
		RedisClient:               redisClient,
	}
}
