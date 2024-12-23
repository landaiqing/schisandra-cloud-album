package sms

import (
	"context"
	gosms "github.com/pkg6/go-sms"
	"github.com/pkg6/go-sms/gateways"
	"github.com/pkg6/go-sms/gateways/aliyun"
	"net/http"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/utils"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendSmsByAliyunLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendSmsByAliyunLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendSmsByAliyunLogic {
	return &SendSmsByAliyunLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendSmsByAliyunLogic) SendSmsByAliyun(req *types.SmsSendRequest) error {
	checkRotateData := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !checkRotateData {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, i18n.FormatText(l.ctx, "sms.verificationFailure")))
	}
	isPhone := utils.IsPhone(req.Phone)
	if !isPhone {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.phoneFormatError"))
	}
	val := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if val != "" {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "sms.smsSendTooFrequently"))
	}
	sms := gosms.NewParser(gateways.Gateways{
		ALiYun: aliyun.ALiYun{
			Host:            l.svcCtx.Config.SMS.Ali.Host,
			AccessKeyId:     l.svcCtx.Config.SMS.Ali.AccessKeyId,
			AccessKeySecret: l.svcCtx.Config.SMS.Ali.AccessKeySecret,
		},
	})
	code := utils.GenValidateCode(6)
	wrong := l.svcCtx.RedisClient.Set(l.ctx, constant.UserSmsRedisPrefix+req.Phone, code, time.Minute).Err()
	if wrong != nil {
		return errors.New(http.StatusInternalServerError, wrong.Error())
	}
	_, err := sms.Send(req.Phone, gosms.MapStringAny{
		"content":  "您的验证码是：****。请不要把验证码泄露给其他人。",
		"template": l.svcCtx.Config.SMS.Ali.TemplateCode,
		"signName": l.svcCtx.Config.SMS.Ali.Signature,
		"data": gosms.MapStrings{
			"code": code,
		},
	}, nil)
	if err != nil {
		return errors.New(http.StatusInternalServerError, err.Error())
	}
	return nil

}
