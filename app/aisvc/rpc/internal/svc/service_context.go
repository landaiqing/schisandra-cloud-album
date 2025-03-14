package svc

import (
	"github.com/Kagami/go-face"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"gocv.io/x/gocv"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/query"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/config"
	"schisandra-album-cloud-microservices/common/caffe_classifier"
	"schisandra-album-cloud-microservices/common/clarity"
	"schisandra-album-cloud-microservices/common/face_recognizer"
	"schisandra-album-cloud-microservices/common/miniox"
	"schisandra-album-cloud-microservices/common/redisx"
	"schisandra-album-cloud-microservices/common/tf_classifier"
)

type ServiceContext struct {
	Config          config.Config
	FaceRecognizer  *face.Recognizer
	DB              *query.Query
	RedisClient     *redis.Client
	TfNet           *gocv.Net
	TfDesc          []string
	CaffeNet        *gocv.Net
	CaffeDesc       []string
	MinioClient     *minio.Client
	ClarityDetector *clarity.ConcurrentDetector
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redisx.NewRedis(c.RedisConf.Host, c.RedisConf.Pass, c.RedisConf.DB)
	_, queryDB := mysql.NewMySQL(c.Mysql.DataSource, c.Mysql.MaxOpenConn, c.Mysql.MaxIdleConn, redisClient)
	tfClassifier, tfDesc := tf_classifier.NewTFClassifier()
	caffeClassifier, caffeDesc := caffe_classifier.NewCaffeClassifier()
	return &ServiceContext{
		Config:         c,
		FaceRecognizer: face_recognizer.NewFaceRecognition(),
		DB:             queryDB,
		RedisClient:    redisClient,
		TfNet:          tfClassifier,
		TfDesc:         tfDesc,
		CaffeNet:       caffeClassifier,
		CaffeDesc:      caffeDesc,
		MinioClient:    miniox.NewMinio(c.Minio.Endpoint, c.Minio.AccessKeyID, c.Minio.SecretAccessKey, c.Minio.UseSSL),
		ClarityDetector: clarity.NewConcurrentDetector(clarity.WithMeanThreshold(13.0), // 提高均值阈值
			clarity.WithLaplaceStdThreshold(25.0), // 提高标准差阈值
			clarity.WithMaxWorkers(8)),
	}
}
