package user

import (
	"context"
	"errors"
	"net/http"
	"schisandra-album-cloud-microservices/common/constant"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	utils2 "schisandra-album-cloud-microservices/common/utils"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
)

type ResetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewResetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ResetPasswordLogic {
	return &ResetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (err error) {
	if !utils2.IsPhone(req.Phone) {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.phoneFormatError"))
	}
	if req.Password != req.Repassword {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.passwordNotMatch"))
	}
	if !utils2.IsPassword(req.Password) {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.passwordFormatError"))
	}
	code := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if code == "" {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.captchaExpired"))
	}
	if req.Captcha != code {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.captchaError"))
	}
	// 验证码检查通过后立即删除或标记为已使用
	if err = l.svcCtx.RedisClient.Del(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Err(); err != nil {
		return errors2.New(http.StatusInternalServerError, err.Error())
	}
	authUser := l.svcCtx.DB.ScaAuthUser

	userInfo, err := authUser.Where(authUser.Phone.Eq(req.Phone)).First()
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.userNotRegistered"))
	}
	if err != nil {
		return errors2.New(http.StatusInternalServerError, err.Error())
	}
	encrypt, err := utils2.Encrypt(req.Password)
	if err != nil {
		return errors2.New(http.StatusInternalServerError, err.Error())
	}

	affected, err := authUser.Where(authUser.ID.Eq(userInfo.ID), authUser.Phone.Eq(req.Phone)).Update(authUser.Password, encrypt)
	if err != nil {
		return errors2.New(http.StatusInternalServerError, err.Error())
	}
	if affected.RowsAffected == 0 {
		return errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.resetPasswordError"))
	}
	return nil
}
