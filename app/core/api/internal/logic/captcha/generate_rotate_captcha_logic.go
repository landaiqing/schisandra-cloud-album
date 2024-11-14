package captcha

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/generate"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateRotateCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateRotateCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateRotateCaptchaLogic {
	return &GenerateRotateCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateRotateCaptchaLogic) GenerateRotateCaptcha() (resp *types.Response, err error) {
	captcha, err := generate.GenerateRotateCaptcha(l.svcCtx.RotateCaptcha, l.svcCtx.RedisClient, l.ctx)
	if err != nil {
		return response.Error(), err
	}
	return response.SuccessWithData(captcha), nil
}
