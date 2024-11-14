package sms

import (
	"context"
	"time"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSmsByTestLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendSmsByTestLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsByTestLogic {
	return &SendSmsByTestLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendSmsByTestLogic) SendSmsByTest(req *types.SmsSendRequest) (resp *types.Response, err error) {
	checkRotateData := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !checkRotateData {
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure", "验证码错误"), nil
	}
	isPhone := utils.IsPhone(req.Phone)
	if !isPhone {
		return response.ErrorWithI18n(l.ctx, "login.phoneFormatError", "手机号格式错误"), nil
	}
	val := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if val != "" {
		return response.ErrorWithI18n(l.ctx, "sms.smsSendTooFrequently", "验证码发送过于频繁，请稍后再试"), nil
	}
	code := utils.GenValidateCode(6)
	wrong := l.svcCtx.RedisClient.Set(l.ctx, constant.UserSmsRedisPrefix+req.Phone, code, time.Minute).Err()
	if wrong != nil {
		return response.ErrorWithI18n(l.ctx, "sms.smsSendFailed", "验证码发送失败"), wrong
	}
	return response.Success(), nil
}
