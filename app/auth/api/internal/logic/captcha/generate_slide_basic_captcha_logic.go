package captcha

import (
	"context"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/common/captcha/generate"
	"schisandra-album-cloud-microservices/common/errors"

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

func (l *GenerateSlideBasicCaptchaLogic) GenerateSlideBasicCaptcha() (resp map[string]interface{}, err error) {
	captcha, err := generate.GenerateSlideBasicCaptcha(l.svcCtx.SlideCaptcha, l.svcCtx.RedisClient, l.ctx)
	if err != nil {
		return nil, errors.New(http.StatusInternalServerError, err.Error())
	}
	return captcha, nil
}
