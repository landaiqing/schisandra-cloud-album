package generate

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha-assets/helper"
	"github.com/wenlng/go-captcha/v2/slide"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

// GenerateSlideBasicCaptcha generate slide basic captcha
func GenerateSlideBasicCaptcha(slide slide.Captcha, redis *redis.Client, ctx context.Context) (map[string]interface{}, error) {
	captData, err := slide.Generate()
	if err != nil {
		return nil, err
	}
	blockData := captData.GetData()
	if blockData == nil {
		return nil, errors.New("block data is nil")
	}
	var masterImageBase64, tileImageBase64 string
	masterImageBase64 = captData.GetMasterImage().ToBase64()

	tileImageBase64 = captData.GetTileImage().ToBase64()

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
		"key":          key,
		"image":        masterImageBase64,
		"thumb":        tileImageBase64,
		"thumb_width":  blockData.Width,
		"thumb_height": blockData.Height,
		"thumb_x":      blockData.TileX,
		"thumb_y":      blockData.TileY,
	}, nil

}
