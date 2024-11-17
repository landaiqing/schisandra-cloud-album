package captcha

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/generate"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateSlideBasicCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateSlideBasicCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateSlideBasicCaptchaLogic {
	return &GenerateSlideBasicCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateSlideBasicCaptchaLogic) GenerateSlideBasicCaptcha() (resp *types.Response, err error) {
	captcha, err := generate.GenerateSlideBasicCaptcha(l.svcCtx.SlideCaptcha, l.svcCtx.RedisClient, l.ctx)
	if err != nil {
		return nil, err
	}
	return response.SuccessWithData(captcha), nil
}
