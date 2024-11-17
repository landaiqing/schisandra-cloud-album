package sms

import (
	"context"
	"time"

	gosms "github.com/pkg6/go-sms"
	"github.com/pkg6/go-sms/gateways"
	"github.com/pkg6/go-sms/gateways/smsbao"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSmsBySmsbaoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendSmsBySmsbaoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsBySmsbaoLogic {
	return &SendSmsBySmsbaoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendSmsBySmsbaoLogic) SendSmsBySmsbao(req *types.SmsSendRequest) (resp *types.Response, err error) {
	checkRotateData := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !checkRotateData {
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure"), nil
	}
	isPhone := utils.IsPhone(req.Phone)
	if !isPhone {
		return response.ErrorWithI18n(l.ctx, "login.phoneFormatError"), nil
	}
	val := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if val != "" {
		return response.ErrorWithI18n(l.ctx, "sms.smsSendTooFrequently"), nil
	}
	sms := gosms.NewParser(gateways.Gateways{
		SmsBao: smsbao.SmsBao{
			User:     l.svcCtx.Config.SMS.SMSBao.Username,
			Password: l.svcCtx.Config.SMS.SMSBao.Password,
		},
	})
	code := utils.GenValidateCode(6)
	wrong := l.svcCtx.RedisClient.Set(l.ctx, constant.UserSmsRedisPrefix+req.Phone, code, time.Minute).Err()
	if wrong != nil {
		return response.ErrorWithI18n(l.ctx, "sms.smsSendFailed"), wrong
	}
	_, err = sms.Send(req.Phone, gosms.MapStringAny{
		"content": "您的验证码是：" + code + "。请不要把验证码泄露给其他人。",
	}, nil)
	if err != nil {
		return response.ErrorWithI18n(l.ctx, "sms.smsSendFailed"), err
	}
	return response.Success(), nil
}
