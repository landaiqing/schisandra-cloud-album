package user

import (
	"context"
	"net/http"
	"strconv"

	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	randomname "schisandra-album-cloud-microservices/app/core/api/common/random_name"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuser"
)

type PhoneLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewPhoneLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *PhoneLoginLogic {
	return &PhoneLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *PhoneLoginLogic) PhoneLogin(r *http.Request, w http.ResponseWriter, req *types.PhoneLoginRequest) (resp *types.Response, err error) {
	if !utils.IsPhone(req.Phone) {
		return response.ErrorWithI18n(l.ctx, "login.phoneFormatError"), nil
	}
	code := l.svcCtx.RedisClient.Get(l.ctx, constant.UserSmsRedisPrefix+req.Phone).Val()
	if code == "" {
		return response.ErrorWithI18n(l.ctx, "login.captchaExpired"), nil
	}
	if req.Captcha != code {
		return response.ErrorWithI18n(l.ctx, "login.captchaError"), nil
	}
	user, err := l.svcCtx.MySQLClient.ScaAuthUser.Query().Where(scaauthuser.Phone(req.Phone), scaauthuser.Deleted(0)).First(l.ctx)
	tx, wrong := l.svcCtx.MySQLClient.Tx(l.ctx)
	if wrong != nil {
		return nil, err
	}
	if ent.IsNotFound(err) {
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		avatar := utils.GenerateAvatar(uidStr)
		name := randomname.GenerateName()

		addUser, fault := l.svcCtx.MySQLClient.ScaAuthUser.Create().
			SetUID(uidStr).
			SetPhone(req.Phone).
			SetAvatar(avatar).
			SetNickname(name).
			SetDeleted(constant.NotDeleted).
			SetGender(constant.Male).
			Save(l.ctx)
		if fault != nil {
			err = tx.Rollback()
			return nil, err
		}
		_, err = l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User)
		if err != nil {
			err = tx.Rollback()
			return nil, err
		}
		data, result := HandleUserLogin(addUser, l.svcCtx, req.AutoLogin, r, w, l.ctx)
		if !result {
			err = tx.Rollback()
			return response.ErrorWithI18n(l.ctx, "login.registerError"), err
		}
		// 记录用户登录设备
		if !GetUserLoginDevice(addUser.UID, r, l.svcCtx.Ip2Region, l.svcCtx.MySQLClient, l.ctx) {
			return response.ErrorWithI18n(l.ctx, "login.registerError"), nil
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
		}
		return response.SuccessWithData(data), nil
	} else if err == nil {
		data, result := HandleUserLogin(user, l.svcCtx, req.AutoLogin, r, w, l.ctx)
		if !result {
			err = tx.Rollback()
			return response.ErrorWithI18n(l.ctx, "login.loginFailed"), err
		}
		// 记录用户登录设备
		if !GetUserLoginDevice(user.UID, r, l.svcCtx.Ip2Region, l.svcCtx.MySQLClient, l.ctx) {
			return response.ErrorWithI18n(l.ctx, "login.loginFailed"), nil
		}
		err = tx.Commit()
		if err != nil {
			tx.Rollback()
		}
		return response.SuccessWithData(data), nil
	} else {
		return response.ErrorWithI18n(l.ctx, "login.loginFailed"), nil
	}
}
