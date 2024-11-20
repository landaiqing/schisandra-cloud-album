package config

import "github.com/zeromicro/go-zero/rest"

type Config struct {
	rest.RestConf
	Web struct {
		URL string
	}
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
	Mongo struct {
		Uri        string
		Username   string
		Password   string
		AuthSource string
		Database   string
	}
	Wechat struct {
		AppID     string
		AppSecret string
		Token     string
		AESKey    string
	}
	OAuth struct {
		Github struct {
			ClientID     string
			ClientSecret string
			RedirectURI  string
		}
		QQ struct {
			ClientID     string
			ClientSecret string
			RedirectURI  string
		}
		Gitee struct {
			ClientID     string
			ClientSecret string
			RedirectURI  string
		}
	}
	SMS struct {
		Ali struct {
			Host            string
			AccessKeyId     string
			AccessKeySecret string
			Signature       string
			TemplateCode    string
		}
		SMSBao struct {
			Username string
			Password string
		}
	}
}
