package wechat_official

import (
	"os"

	"github.com/ArtisanCloud/PowerWeChat/v3/src/kernel"
	"github.com/ArtisanCloud/PowerWeChat/v3/src/officialAccount"
)

// NewWechatPublic 微信公众号实例化
func NewWechatPublic(appId, appSecret, token, aesKey, addr, pass string, db int) *officialAccount.OfficialAccount {
	OfficialAccountApp, err := officialAccount.NewOfficialAccount(&officialAccount.UserConfig{
		AppID:  appId,
		Secret: appSecret,
		Token:  token,
		AESKey: aesKey,
		Log: officialAccount.Log{
			Level:  "error",
			File:   "/logs/wechat/wechat_official.log",
			Stdout: false,
		},
		ResponseType: os.Getenv("response_type"),
		HttpDebug:    false,
		Debug:        false,
		Cache: kernel.NewRedisClient(&kernel.UniversalOptions{
			Addrs:    []string{addr},
			Password: pass,
			DB:       db,
		}),
	})
	if err != nil {
		panic(err)
	}
	return OfficialAccountApp
}
