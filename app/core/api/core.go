package main

import (
	"flag"
	"fmt"

	"schisandra-album-cloud-microservices/app/core/api/common/middleware"
	"schisandra-album-cloud-microservices/app/core/api/internal/config"
	"schisandra-album-cloud-microservices/app/core/api/internal/handler"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/core.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, rest.WithCors())
	defer server.Stop()
	// i18n middleware
	server.Use(middleware.I18nMiddleware)
	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
