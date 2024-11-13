package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/slide"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

// VerifySlideCaptcha verify slide captcha
func VerifySlideCaptcha(context context.Context, redis *redis.Client, point []int64, key string) bool {
	cacheDataByte := redis.Get(context, constant.UserCaptchaPrefix+key).Val()
	if len(cacheDataByte) == 0 {
		return false
	}
	var dct *slide.Block
	if err := json.Unmarshal([]byte(cacheDataByte), &dct); err != nil {
		return false
	}

	chkRet := false
	if 2 == len(point) {
		sx, _ := strconv.ParseFloat(fmt.Sprintf("%v", point[0]), 64)
		sy, _ := strconv.ParseFloat(fmt.Sprintf("%v", point[1]), 64)
		chkRet = slide.CheckPoint(int64(sx), int64(sy), int64(dct.X), int64(dct.Y), 4)
	}
	if chkRet {
		return true
	}
	return false
}
