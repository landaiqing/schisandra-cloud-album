package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
	}
	Mysql struct {
		DataSource  string
		MaxOpenConn int
		MaxIdleConn int
	}
	Redis struct {
		Host string
		Pass string
		DB   int
	}
}
