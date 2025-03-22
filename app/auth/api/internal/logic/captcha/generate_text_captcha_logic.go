package captcha

import (
	"context"
	"net/http"
	"schisandra-album-cloud-microservices/common/captcha/generate"
	"schisandra-album-cloud-microservices/common/errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GenerateTextCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGenerateTextCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GenerateTextCaptchaLogic {
	return &GenerateTextCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GenerateTextCaptchaLogic) GenerateTextCaptcha() (resp *types.TextCaptchaResponse, err error) {
	captcha, err := generate.GenerateBasicTextCaptcha(l.svcCtx.TextCaptcha, l.svcCtx.RedisClient, l.ctx)
	if err != nil {
		return nil, errors.New(http.StatusInternalServerError, err.Error())
	}
	return &types.TextCaptchaResponse{
		Key:   captcha["key"].(string),
		Image: captcha["image"].(string),
		Thumb: captcha["thumb"].(string),
	}, nil
}
