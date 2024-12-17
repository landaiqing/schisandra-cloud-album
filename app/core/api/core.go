package main

import (
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"

	"schisandra-album-cloud-microservices/app/core/api/common/middleware"
	"schisandra-album-cloud-microservices/app/core/api/internal/config"
	"schisandra-album-cloud-microservices/app/core/api/internal/handler"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/repository/idgenerator"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCustomCors(middleware.CORSMiddleware(), nil),
		rest.WithUnauthorizedCallback(middleware.UnauthorizedCallbackMiddleware()))
	defer server.Stop()
	// i18n middleware
	server.Use(middleware.I18nMiddleware)
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// init id generator
	idgenerator.NewIDGenerator()
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
