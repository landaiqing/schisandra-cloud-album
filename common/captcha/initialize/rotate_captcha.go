package initialize

import (
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/rotate"
)

// NewRotateCaptcha 初始化旋转验证码
func NewRotateCaptcha() rotate.Captcha {
	builder := rotate.NewBuilder(rotate.WithRangeAnglePos([]option.RangeVal{
		{Min: 20, Max: 330},
	}))

	// background images
	imgs, err := images.GetImages()
	if err != nil {
		panic(err)
	}

	// set resources
	builder.SetResources(
		rotate.WithImages(imgs),
	)
	return builder.Make()
}
