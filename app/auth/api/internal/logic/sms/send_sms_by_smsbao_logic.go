package sms

import (
	"context"
	"net/http"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	utils2 "schisandra-album-cloud-microservices/common/utils"
	"time"

	gosms "github.com/pkg6/go-sms"
	"github.com/pkg6/go-sms/gateways"
	"github.com/pkg6/go-sms/gateways/smsbao"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

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

func (l *SendSmsBySmsbaoLogic) SendSmsBySmsbao(req *types.SmsSendRequest) (err error) {
	checkRotateData := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !checkRotateData {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
	}
	isPhone := utils2.IsPhone(req.Phone)
	if !isPhone {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.phoneFormatError"))
	}
	val := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if val != "" {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "sms.smsSendTooFrequently"))
	}
	sms := gosms.NewParser(gateways.Gateways{
		SmsBao: smsbao.SmsBao{
			User:     l.svcCtx.Config.SMS.SMSBao.Username,
			Password: l.svcCtx.Config.SMS.SMSBao.Password,
		},
	})
	code := utils2.GenValidateCode(6)
	wrong := l.svcCtx.RedisClient.Set(l.ctx, constant.UserSmsRedisPrefix+req.Phone, code, time.Minute).Err()
	if wrong != nil {
		return errors.New(http.StatusInternalServerError, wrong.Error())
	}
	_, err = sms.Send(req.Phone, gosms.MapStringAny{
		"content": "您的验证码是：" + code + "。请不要把验证码泄露给其他人。",
	}, nil)
	if err != nil {
		return errors.New(http.StatusInternalServerError, err.Error())
	}
	return nil
}
