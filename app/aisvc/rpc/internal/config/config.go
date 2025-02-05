package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Mysql struct {
		DataSource  string
		MaxOpenConn int
		MaxIdleConn int
	}
	RedisConf struct {
		Host string
		Pass string
		DB   int
	}
	Minio struct {
		Endpoint        string
		AccessKeyID     string
		SecretAccessKey string
		UseSSL          bool
	}
}
