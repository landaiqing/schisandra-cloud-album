package user

import (
	"context"
	"errors"
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
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
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
	authUser := model.ScaAuthUser{
		Phone:   req.Phone,
		Deleted: constant.NotDeleted,
	}
	has, err := l.svcCtx.DB.Get(&authUser)
	if err != nil {
		return nil, err
	}
	tx := l.svcCtx.DB.NewSession()
	defer tx.Close()
	if err = tx.Begin(); err != nil {
		return nil, err
	}

	if !has {
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		avatar := utils.GenerateAvatar(uidStr)
		name := randomname.GenerateName()

		user := model.ScaAuthUser{
			UID:      uidStr,
			Phone:    req.Phone,
			Avatar:   avatar,
			Nickname: name,
			Deleted:  constant.NotDeleted,
			Gender:   constant.Male,
		}
		insert, err := tx.Insert(&user)
		if err != nil || insert == 0 {
			return nil, errors.New("register failed")
		}
		_, err = l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User)
		if err != nil {
			return nil, err
		}
		data, err := HandleUserLogin(user, l.svcCtx, req.AutoLogin, r, w, l.ctx)
		if err != nil {
			return nil, err
		}
		// 记录用户登录设备
		if err = GetUserLoginDevice(user.UID, r, l.svcCtx.Ip2Region, l.svcCtx.DB, l.ctx); err != nil {
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return response.SuccessWithData(data), nil
	} else {
		data, err := HandleUserLogin(authUser, l.svcCtx, req.AutoLogin, r, w, l.ctx)
		if err != nil {
			return nil, err
		}
		// 记录用户登录设备
		if err = GetUserLoginDevice(authUser.UID, r, l.svcCtx.Ip2Region, l.svcCtx.DB, l.ctx); err != nil {
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return response.SuccessWithData(data), nil
	}
}
