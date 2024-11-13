package svc

import (
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"

	"github.com/zeromicro/go-zero/rest"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"schisandra-album-cloud-microservices/app/core/api/internal/config"
	"schisandra-album-cloud-microservices/app/core/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/core/api/repository/ip2region"
	"schisandra-album-cloud-microservices/app/core/api/repository/mongodb"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/redis_session"
	"schisandra-album-cloud-microservices/app/core/api/repository/redisx"
)

type ServiceContext struct {
	Config                    config.Config
	SecurityHeadersMiddleware rest.Middleware
	MySQLClient               *ent.Client
	RedisClient               *redis.Client
	MongoClient               *mongo.Database
	Session                   *redisstore.RedisStore
	Ip2Region                 *xdb.Searcher
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                    c,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		MySQLClient:               mysql.NewMySQL(c.Mysql.DataSource),
		RedisClient:               redisx.NewRedis(c.Redis.Host, c.Redis.Pass, c.Redis.DB),
		MongoClient:               mongodb.NewMongoDB(c.Mongo.Uri, c.Mongo.Username, c.Mongo.Password, c.Mongo.AuthSource, c.Mongo.Database),
		Session:                   redis_session.NewRedisSession(c.Redis.Host, c.Redis.Pass),
		Ip2Region:                 ip2region.NewIP2Region(),
	}
}
