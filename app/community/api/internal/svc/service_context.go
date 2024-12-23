package svc

import (
	"github.com/casbin/casbin/v2"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/rest"
	sensitive "github.com/zmexing/go-sensitive-word"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"schisandra-album-cloud-microservices/app/community/api/internal/config"
	"schisandra-album-cloud-microservices/app/community/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/community/api/model/mongodb"
	"schisandra-album-cloud-microservices/app/community/api/model/mysql"
	"schisandra-album-cloud-microservices/app/community/api/model/mysql/query"
	"schisandra-album-cloud-microservices/common/casbinx"
	"schisandra-album-cloud-microservices/common/ip2region"
	"schisandra-album-cloud-microservices/common/redisx"
	"schisandra-album-cloud-microservices/common/sensitivex"
)

type ServiceContext struct {
	Config                    config.Config
	SecurityHeadersMiddleware rest.Middleware
	CasbinVerifyMiddleware    rest.Middleware
	AuthorizationMiddleware   rest.Middleware
	NonceMiddleware           rest.Middleware
	MongoClient               *mongo.Database
	DB                        *query.Query
	CasbinEnforcer            *casbin.SyncedCachedEnforcer
	RedisClient               *redis.Client
	Sensitive                 *sensitive.Manager
	Ip2Region                 *xdb.Searcher
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redisx.NewRedis(c.Redis.Host, c.Redis.Pass, c.Redis.DB)
	db, queryDB := mysql.NewMySQL(c.Mysql.DataSource, c.Mysql.MaxOpenConn, c.Mysql.MaxIdleConn, redisClient)
	casbinEnforcer := casbinx.NewCasbin(db)
	return &ServiceContext{
		Config:                    c,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		CasbinVerifyMiddleware:    middleware.NewCasbinVerifyMiddleware(casbinEnforcer).Handle,
		AuthorizationMiddleware:   middleware.NewAuthorizationMiddleware().Handle,
		NonceMiddleware:           middleware.NewNonceMiddleware(redisClient).Handle,
		MongoClient:               mongodb.NewMongoDB(c.Mongo.Uri, c.Mongo.Username, c.Mongo.Password, c.Mongo.AuthSource, c.Mongo.Database),
		DB:                        queryDB,
		CasbinEnforcer:            casbinEnforcer,
		RedisClient:               redisClient,
		Sensitive:                 sensitivex.NewSensitive(),
		Ip2Region:                 ip2region.NewIP2Region(),
	}
}
