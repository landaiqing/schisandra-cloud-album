package verify

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/redis/go-redis/v9"
	"github.com/wenlng/go-captcha/v2/click"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
)

// VerifyBasicTextCaptcha verify basic text captcha
func VerifyBasicTextCaptcha(dots string, key string, redis *redis.Client, ctx context.Context) bool {
	cacheDataByte, err := redis.Get(ctx, constant.UserCaptchaPrefix+key).Bytes()
	if len(cacheDataByte) == 0 || err != nil {
		return false
	}
	src := strings.Split(dots, ",")

	var dct map[int]*click.Dot
	if err := json.Unmarshal(cacheDataByte, &dct); err != nil {

		return false
	}
	chkRet := false
	if (len(dct) * 2) == len(src) {
		for i := 0; i < len(dct); i++ {
			dot := dct[i]
			j := i * 2
			k := i*2 + 1
			sx, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[j]), 64)
			sy, _ := strconv.ParseFloat(fmt.Sprintf("%v", src[k]), 64)

			chkRet = click.CheckPoint(int64(sx), int64(sy), int64(dot.X), int64(dot.Y), int64(dot.Width), int64(dot.Height), 0)
			if !chkRet {
				break
			}
		}
	}
	if chkRet {
		return true
	}
	return false
}
