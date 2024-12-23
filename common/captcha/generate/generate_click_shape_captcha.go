package generate

import (
	"context"
	"encoding/json"
	"errors"
	"schisandra-album-cloud-microservices/common/constant"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha-assets/helper"
	"github.com/wenlng/go-captcha/v2/click"
)

// GenerateClickShapeCaptcha generate click shape captcha
func GenerateClickShapeCaptcha(click click.Captcha, redis redis.Client, ctx context.Context) (map[string]interface{}, error) {
	captData, err := click.Generate()
	if err != nil {
		return nil, err
	}
	dotData := captData.GetData()
	if dotData == nil {
		return nil, errors.New("captcha data is nil")
	}
	var masterImageBase64, thumbImageBase64 string
	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}
	thumbImageBase64, err = captData.GetThumbImage().ToBase64()
	if err != nil {
		return nil, err
	}

	dotsByte, err := json.Marshal(dotData)
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
