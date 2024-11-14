package generate

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha-assets/helper"
	"github.com/wenlng/go-captcha/v2/click"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

// GenerateBasicTextCaptcha generates a basic text captcha and saves it to redis.
func GenerateBasicTextCaptcha(capt click.Captcha, redis redis.Client, ctx context.Context) map[string]interface{} {
	captData, err := capt.Generate()
	if err != nil {
		return nil
	}
	dotData := captData.GetData()
	if dotData == nil {
		return nil
	}
	var masterImageBase64, thumbImageBase64 string
	masterImageBase64 = captData.GetMasterImage().ToBase64()
	thumbImageBase64 = captData.GetThumbImage().ToBase64()

	dotsByte, err := json.Marshal(dotData)
	if err != nil {
		return nil
	}
	key := helper.StringToMD5(string(dotsByte))
	err = redis.Set(ctx, constant.UserCaptchaPrefix+key, dotsByte, time.Minute).Err()
	if err != nil {
		return nil
	}
	return map[string]interface{}{
		"key":   key,
		"image": masterImageBase64,
		"thumb": thumbImageBase64,
	}

}
