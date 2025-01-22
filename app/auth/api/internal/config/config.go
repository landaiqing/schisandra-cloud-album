package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	AiSvcRpc zrpc.RpcClientConf
	Web      struct {
		URL string
	}
	Auth struct {
		AccessSecret string
	}
	Encrypt struct {
		Key string
		IV  string
	}
	Mysql struct {
		DataSource  string
		MaxOpenConn int
		MaxIdleConn int
	}
	Mongo struct {
		Uri        string
		Username   string
		Password   string
		AuthSource string
		Database   string
	}
	Redis struct {
		Host string
		Pass string
		DB   int
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
	Map struct {
		Key string
	}
}
