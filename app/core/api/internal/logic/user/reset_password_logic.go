package user

import (
	"context"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuser"

	"github.com/zeromicro/go-zero/core/logx"
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
		return response.ErrorWithI18n(l.ctx, "login.captchaError"), err
	}
	user, err := l.svcCtx.MySQLClient.ScaAuthUser.Query().Where(scaauthuser.Phone(req.Phone), scaauthuser.Deleted(constant.NotDeleted)).First(l.ctx)
	if err != nil && ent.IsNotFound(err) {
		return response.ErrorWithI18n(l.ctx, "login.userNotRegistered"), err
	}
	encrypt, err := utils.Encrypt(req.Password)
	if err != nil {
		return response.ErrorWithI18n(l.ctx, "login.resetPasswordError"), err
	}
	err = user.Update().SetPassword(encrypt).Exec(l.ctx)
	if err != nil {
		return response.ErrorWithI18n(l.ctx, "login.resetPasswordError"), err
	}
	return response.Success(), nil
}
