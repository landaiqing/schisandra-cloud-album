package sms

import (
	"context"
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

func (l *SendSmsByTestLogic) SendSmsByTest(req *types.SmsSendRequest) (err error) {
	checkRotateData := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !checkRotateData {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
	}
	isPhone := utils.IsPhone(req.Phone)
	if !isPhone {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.phoneFormatError"))
	}
	val := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if val != "" {
		return errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "sms.smsSendTooFrequently"))
	}
	code := utils.GenValidateCode(6)
	wrong := l.svcCtx.RedisClient.Set(l.ctx, constant.UserSmsRedisPrefix+req.Phone, code, time.Minute).Err()
	if wrong != nil {
		return errors.New(http.StatusInternalServerError, wrong.Error())
	}
	return nil
}
