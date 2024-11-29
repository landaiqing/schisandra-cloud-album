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

// GenerateSlideRegionCaptcha generate slide region captcha
func GenerateSlideRegionCaptcha(slide slide.Captcha, redis redis.Client, ctx context.Context) (map[string]interface{}, error) {
	captData, err := slide.Generate()
	if err != nil {
		return nil, err
	}

	blockData := captData.GetData()
	if blockData == nil {
		return nil, errors.New("block data is nil")
	}

	var masterImageBase64, tileImageBase64 string
	masterImageBase64, err = captData.GetMasterImage().ToBase64()
	if err != nil {
		return nil, err
	}
	tileImageBase64, err = captData.GetTileImage().ToBase64()
	if err != nil {
		return nil, err
	}

	blockByte, err := json.Marshal(blockData)
	if err != nil {
		return nil, err
	}
	key := helper.StringToMD5(string(blockByte))
	err = redis.Set(ctx, constant.UserCaptchaPrefix+key, blockByte, time.Minute).Err()
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{
		"code":        0,
		"key":         key,
		"image":       masterImageBase64,
		"tile":        tileImageBase64,
		"tile_width":  blockData.Width,
		"tile_height": blockData.Height,
		"tile_x":      blockData.TileX,
		"tile_y":      blockData.TileY,
	}, nil
}
