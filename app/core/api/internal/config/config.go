package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Auth struct {
		AccessSecret string
		AccessExpire int64
	}
	Mysql struct {
		DataSource string
	}
	Redis struct {
		Host string
		Pass string
		DB   int
	}
	Mongo struct {
		Uri        string
		Username   string
		Password   string
		AuthSource string
		Database   string
	}
}
