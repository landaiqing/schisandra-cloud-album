package captcha

import (
	"github.com/wenlng/go-captcha-assets/resources/images"
	"github.com/wenlng/go-captcha-assets/resources/shapes"
	"github.com/wenlng/go-captcha/v2/base/option"
	"github.com/wenlng/go-captcha/v2/click"
)

// NewClickShapeCaptcha 初始化点击形状验证码
func NewClickShapeCaptcha() click.Captcha {
	builder := click.NewBuilder(
		click.WithRangeLen(option.RangeVal{Min: 3, Max: 6}),
		click.WithRangeVerifyLen(option.RangeVal{Min: 2, Max: 3}),
		click.WithRangeThumbBgDistort(1),
		click.WithIsThumbNonDeformAbility(true),
	)

	// shape
	// click.WithUseShapeOriginalColor(false) -> Random rewriting of graphic colors
	shapeMaps, err := shapes.GetShapes()
	if err != nil {
		panic(err)
	}

	// background images
	imgs, err := images.GetImages()
	if err != nil {
		panic(err)
	}

	// set resources
	builder.SetResources(
		click.WithShapes(shapeMaps),
		click.WithBackgrounds(imgs),
	)
	return builder.MakeWithShape()
}
