package user

import (
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
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

func (l *ResetPasswordLogic) ResetPassword(req *types.ResetPasswordRequest) (resp *types.Response, err error) {
	if !utils.IsPhone(req.Phone) {
		return response.ErrorWithI18n(l.ctx, "login.phoneFormatError"), nil
	}
	if req.Password != req.Repassword {
		return response.ErrorWithI18n(l.ctx, "login.passwordNotMatch"), nil
	}
	if !utils.IsPassword(req.Password) {
		return response.ErrorWithI18n(l.ctx, "login.passwordFormatError"), nil
	}
	code := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if code == "" {
		return response.ErrorWithI18n(l.ctx, "login.captchaExpired"), nil
	}
	if req.Captcha != code {
		return response.ErrorWithI18n(l.ctx, "login.captchaError"), nil
	}
	// 验证码检查通过后立即删除或标记为已使用
	if err = l.svcCtx.RedisClient.Del(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Err(); err != nil {
		return nil, err
	}
	authUser := l.svcCtx.DB.ScaAuthUser

	userInfo, err := authUser.Where(authUser.Phone.Eq(req.Phone)).First()
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return response.ErrorWithI18n(l.ctx, "login.userNotRegistered"), nil
	}
	if err != nil {
		return nil, err
	}
	encrypt, err := utils.Encrypt(req.Password)
	if err != nil {
		return nil, err
	}

	affected, err := authUser.Where(authUser.ID.Eq(userInfo.ID), authUser.Phone.Eq(req.Phone)).Update(authUser.Password, encrypt)
	if err != nil {
		return nil, err
	}
	if affected.RowsAffected == 0 {
		return response.ErrorWithI18n(l.ctx, "login.resetPasswordError"), nil
	}
	return response.Success(), nil
}
