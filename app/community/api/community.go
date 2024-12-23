package main

import (
	"flag"
	"fmt"
	"schisandra-album-cloud-microservices/app/community/api/internal/config"
	"schisandra-album-cloud-microservices/app/community/api/internal/handler"
	"schisandra-album-cloud-microservices/app/community/api/internal/svc"
	"schisandra-album-cloud-microservices/common/idgenerator"
	"schisandra-album-cloud-microservices/common/middleware"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/community.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf,
		rest.WithCustomCors(middleware.CORSMiddleware(), nil),
		rest.WithUnauthorizedCallback(middleware.UnauthorizedCallbackMiddleware()),
		rest.WithUnsignedCallback(middleware.UnsignedCallbackMiddleware()))
	defer server.Stop()

	server.Use(middleware.I18nMiddleware)
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	idgenerator.NewIDGenerator(1)
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
