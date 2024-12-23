package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"schisandra-album-cloud-microservices/common/constant"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/rotate"
)

// VerifyRotateCaptcha verify rotate captcha
func VerifyRotateCaptcha(context context.Context, redis *redis.Client, angle int64, key string) bool {
	cacheDataByte := redis.Get(context, constant.UserCaptchaPrefix+key).Val()
	if len(cacheDataByte) == 0 {
		return false
	}
	var dct *rotate.Block
	if err := json.Unmarshal([]byte(cacheDataByte), &dct); err != nil {
		return false
	}
	sAngle, err := strconv.ParseFloat(fmt.Sprintf("%v", angle), 64)
	if err != nil {
		return false
	}
	chkRet := rotate.CheckAngle(int64(sAngle), int64(dct.Angle), 2)
	if chkRet {
		return true
	}
	return false

}
