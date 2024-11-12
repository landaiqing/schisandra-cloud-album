package svc

import (
	"github.com/zeromicro/go-zero/rest"

	"schisandra-album-cloud-microservices/app/auth/internal/config"
	"schisandra-album-cloud-microservices/app/auth/internal/middleware"
	"schisandra-album-cloud-microservices/common/core"
	"schisandra-album-cloud-microservices/common/ent"
)

type ServiceContext struct {
	Config                    config.Config
	I18nMiddleware            rest.Middleware
	SecurityHeadersMiddleware rest.Middleware
	DB                        *ent.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                    c,
		I18nMiddleware:            middleware.NewI18nMiddleware().Handle,
		SecurityHeadersMiddleware: middleware.NewSecurityHeadersMiddleware().Handle,
		DB:                        core.InitMySQL(c.Mysql.DataSource),
	}
}
