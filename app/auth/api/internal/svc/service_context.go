package svc

import (
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/casbin/casbin/v2"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
	"github.com/zeromicro/go-zero/rest"
	sensitive "github.com/zmexing/go-sensitive-word"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"schisandra-album-cloud-microservices/app/auth/api/internal/config"
	"schisandra-album-cloud-microservices/app/auth/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/auth/model/mongodb"
	"schisandra-album-cloud-microservices/app/auth/model/mysql"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/captcha/initialize"
	"schisandra-album-cloud-microservices/common/casbinx"
	"schisandra-album-cloud-microservices/common/ip2region"
	"schisandra-album-cloud-microservices/common/redisx"
	"schisandra-album-cloud-microservices/common/sensitivex"
	"schisandra-album-cloud-microservices/common/wechat_official"
)

type ServiceContext struct {
	Config                    config.Config
	SecurityHeadersMiddleware rest.Middleware
	CasbinVerifyMiddleware    rest.Middleware
	AuthorizationMiddleware   rest.Middleware
	NonceMiddleware           rest.Middleware
	DB                        *query.Query
	RedisClient               *redis.Client
	Ip2Region                 *xdb.Searcher
	CasbinEnforcer            *casbin.SyncedCachedEnforcer
	WechatOfficial            *officialAccount.OfficialAccount
	MongoClient               *mongo.Database
	RotateCaptcha             rotate.Captcha
	SlideCaptcha              slide.Captcha
	Sensitive                 *sensitive.Manager
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
		DB:                        queryDB,
		RedisClient:               redisClient,
		Ip2Region:                 ip2region.NewIP2Region(),
		CasbinEnforcer:            casbinEnforcer,
		WechatOfficial:            wechat_official.NewWechatPublic(c.Wechat.AppID, c.Wechat.AppSecret, c.Wechat.Token, c.Wechat.AESKey, c.Redis.Host, c.Redis.Pass, c.Redis.DB),
		RotateCaptcha:             initialize.NewRotateCaptcha(),
		SlideCaptcha:              initialize.NewSlideCaptcha(),
		MongoClient:               mongodb.NewMongoDB(c.Mongo.Uri, c.Mongo.Username, c.Mongo.Password, c.Mongo.AuthSource, c.Mongo.Database),
		Sensitive:                 sensitivex.NewSensitive(),
	}
}
