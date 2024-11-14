package generate

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha-assets/helper"
	"github.com/wenlng/go-captcha/v2/rotate"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

// GenerateRotateCaptcha generate rotate captcha
func GenerateRotateCaptcha(captcha rotate.Captcha, redis *redis.Client, ctx context.Context) (map[string]interface{}, error) {
	captchaData, err := captcha.Generate()
	if err != nil {
		return nil, err
	}
	blockData := captchaData.GetData()
	if blockData == nil {
		return nil, errors.New("captcha data is nil")
	}
	masterImageBase64 := captchaData.GetMasterImage().ToBase64()
	thumbImageBase64 := captchaData.GetThumbImage().ToBase64()
	dotsByte, err := json.Marshal(blockData)
	if err != nil {
		return nil, err
	}
	key := helper.StringToMD5(string(dotsByte))
	err = redis.Set(ctx, constant.UserCaptchaPrefix+key, dotsByte, time.Minute).Err()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"key":   key,
		"image": masterImageBase64,
		"thumb": thumbImageBase64,
	}, nil
}
