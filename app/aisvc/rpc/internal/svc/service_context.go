package svc

import (
	"github.com/Kagami/go-face"
	"github.com/redis/go-redis/v9"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/query"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/config"
	"schisandra-album-cloud-microservices/common/face_recognizer"
	"schisandra-album-cloud-microservices/common/redisx"
)

type ServiceContext struct {
	Config         config.Config
	FaceRecognizer *face.Recognizer
	DB             *query.Query
	RedisClient    *redis.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redisx.NewRedis(c.RedisConf.Host, c.RedisConf.Pass, c.RedisConf.DB)
	_, queryDB := mysql.NewMySQL(c.Mysql.DataSource, c.Mysql.MaxOpenConn, c.Mysql.MaxIdleConn, redisClient)
	return &ServiceContext{
		Config:         c,
		FaceRecognizer: face_recognizer.NewFaceRecognition(),
		DB:             queryDB,
		RedisClient:    redisClient,
	}
}
