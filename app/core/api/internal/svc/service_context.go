package svc

import (
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
	"github.com/casbin/casbin/v2"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/rbcervilla/redisstore/v9"
	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/rotate"
	"github.com/wenlng/go-captcha/v2/slide"
	sensitive "github.com/zmexing/go-sensitive-word"

	"github.com/zeromicro/go-zero/rest"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"schisandra-album-cloud-microservices/app/core/api/internal/config"
	"schisandra-album-cloud-microservices/app/core/api/internal/middleware"
	"schisandra-album-cloud-microservices/app/core/api/repository/captcha"
	"schisandra-album-cloud-microservices/app/core/api/repository/casbinx"
	"schisandra-album-cloud-microservices/app/core/api/repository/ip2region"
	"schisandra-album-cloud-microservices/app/core/api/repository/mongodb"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/redis_session"
	"schisandra-album-cloud-microservices/app/core/api/repository/redisx"
	"schisandra-album-cloud-microservices/app/core/api/repository/sensitivex"
	"schisandra-album-cloud-microservices/app/core/api/repository/wechat_public"
)

type ServiceContext struct {
	Config                    config.Config
	SecurityHeadersMiddleware rest.Middleware
	MySQLClient               *ent.Client
	RedisClient               *redis.Client
	MongoClient               *mongo.Database
	Session                   *redisstore.RedisStore
	Ip2Region                 *xdb.Searcher
	CasbinEnforcer            *casbin.CachedEnforcer
	WechatPublic              *officialAccount.OfficialAccount
	Sensitive                 *sensitive.Manager
	RotateCaptcha             rotate.Captcha
	SlideCaptcha              slide.Captcha
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
		CasbinEnforcer:            casbinx.NewCasbin(c.Mysql.DataSource),
		WechatPublic:              wechat_public.NewWechatPublic(c.Wechat.AppID, c.Wechat.AppSecret, c.Wechat.Token, c.Wechat.AESKey, c.Redis.Host, c.Redis.Pass, c.Redis.DB),
		Sensitive:                 sensitivex.NewSensitive(),
		RotateCaptcha:             captcha.NewRotateCaptcha(),
		SlideCaptcha:              captcha.NewSlideCaptcha(),
	}
}
