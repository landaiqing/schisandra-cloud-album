package svc

import (
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/casbin/casbin/v2"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/minio/minio-go/v7"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
	sensitive "github.com/zmexing/go-sensitive-word"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/client/aiservice"
	"schisandra-album-cloud-microservices/app/auth/api/internal/config"
	"schisandra-album-cloud-microservices/app/auth/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/auth/model/mysql"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/captcha/initialize"
	"schisandra-album-cloud-microservices/common/casbinx"
	"schisandra-album-cloud-microservices/common/geo_json"
	"schisandra-album-cloud-microservices/common/ip2region"
	"schisandra-album-cloud-microservices/common/miniox"
	"schisandra-album-cloud-microservices/common/nsqx"
	"schisandra-album-cloud-microservices/common/redisx"
	"schisandra-album-cloud-microservices/common/sensitivex"
	"schisandra-album-cloud-microservices/common/storage"
	"schisandra-album-cloud-microservices/common/storage/manager"
	"schisandra-album-cloud-microservices/common/wechat_official"
	"schisandra-album-cloud-microservices/common/zincx"
)

type ServiceContext struct {
	Config                    config.Config
	AiSvcRpc                  aiservice.AiService
	SecurityHeadersMiddleware rest.Middleware
	CasbinVerifyMiddleware    rest.Middleware
	NonceMiddleware           rest.Middleware
	DB                        *query.Query
	RedisClient               *redis.Client
	Ip2Region                 *xdb.Searcher
	CasbinEnforcer            *casbin.SyncedCachedEnforcer
	WechatOfficial            *officialAccount.OfficialAccount
	RotateCaptcha             rotate.Captcha
	SlideCaptcha              slide.Captcha
	Sensitive                 *sensitive.Manager
	StorageManager            *manager.Manager
	MinioClient               *minio.Client
	GeoRegionData             *geo_json.RegionData
	NSQProducer               *nsq.Producer
	ZincClient                *zincx.ZincClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	redisClient := redisx.NewRedis(c.Redis.Host, c.Redis.Pass, c.Redis.DB)
	db, queryDB := mysql.NewMySQL(c.Mysql.DataSource, c.Mysql.MaxOpenConn, c.Mysql.MaxIdleConn, redisClient)
	casbinEnforcer := casbinx.NewCasbin(db)
	serviceContext := &ServiceContext{
		Config:                    c,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		CasbinVerifyMiddleware:    middleware.NewCasbinVerifyMiddleware(casbinEnforcer).Handle,
		NonceMiddleware:           middleware.NewNonceMiddleware(redisClient).Handle,
		DB:                        queryDB,
		RedisClient:               redisClient,
		Ip2Region:                 ip2region.NewIP2Region(),
		CasbinEnforcer:            casbinEnforcer,
		WechatOfficial:            wechat_official.NewWechatPublic(c.Wechat.AppID, c.Wechat.AppSecret, c.Wechat.Token, c.Wechat.AESKey, c.Redis.Host, c.Redis.Pass, c.Redis.DB),
		RotateCaptcha:             initialize.NewRotateCaptcha(),
		SlideCaptcha:              initialize.NewSlideCaptcha(),
		Sensitive:                 sensitivex.NewSensitive(),
		StorageManager:            storage.InitStorageManager(),
		AiSvcRpc:                  aiservice.NewAiService(zrpc.MustNewClient(c.AiSvcRpc)),
		MinioClient:               miniox.NewMinio(c.Minio.Endpoint, c.Minio.AccessKeyID, c.Minio.SecretAccessKey, c.Minio.UseSSL),
		GeoRegionData:             geo_json.NewGeoJSON(),
		NSQProducer:               nsqx.NewNsqProducer(c.NSQ.NSQDHost),
		ZincClient:                zincx.NewZincClient(c.Zinc.URL, c.Zinc.Username, c.Zinc.Password),
	}
	return serviceContext
}
