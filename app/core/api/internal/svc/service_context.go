package svc

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"

	"schisandra-album-cloud-microservices/app/core/api/internal/config"
	"schisandra-album-cloud-microservices/app/core/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/core/api/repository/mongodb"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql"
)

type ServiceContext struct {
	Config                    config.Config
	SecurityHeadersMiddleware rest.Middleware
	MySQLClient               *ent.Client
	RedisClient               *redis.Redis
	MongoClient               *mongo.Database
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                    c,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		MySQLClient:               mysql.NewMySQL(c.Mysql.DataSource),
		RedisClient: redis.MustNewRedis(redis.RedisConf{
			Host: c.Redis.Host,
			Pass: c.Redis.Pass,
			Type: c.Redis.Type,
			Tls:  c.Redis.Tls,
		}),
		MongoClient: mongodb.NewMongoDB(c.Mongo.Uri, c.Mongo.Username, c.Mongo.Password, c.Mongo.AuthSource, c.Mongo.Database),
	}
}
