package main

import (
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"schisandra-album-cloud-microservices/app/auth/api/internal/mq"
	"schisandra-album-cloud-microservices/common/idgenerator"
	"schisandra-album-cloud-microservices/common/middleware"

	"schisandra-album-cloud-microservices/app/auth/api/internal/config"
	"schisandra-album-cloud-microservices/app/auth/api/internal/handler"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
)

var configFile = flag.String("f", "api/etc/auth.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(
		c.RestConf,
		rest.WithCustomCors(middleware.CORSMiddleware(), nil),
		rest.WithUnauthorizedCallback(middleware.UnauthorizedCallbackMiddleware()),
		rest.WithUnsignedCallback(middleware.UnsignedCallbackMiddleware()))
	defer server.Stop()
	// i18n middleware
	server.Use(middleware.I18nMiddleware)
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// start image process consumer
	go mq.NewImageProcessConsumer(ctx)
	// initialize id generator
	idgenerator.NewIDGenerator(0)
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
