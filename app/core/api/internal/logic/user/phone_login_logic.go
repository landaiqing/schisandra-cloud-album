package user

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

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
	authUser := l.svcCtx.DB.ScaAuthUser
	userInfo, err := authUser.Where(authUser.Phone.Eq(req.Phone)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	tx := l.svcCtx.DB.Begin()
	defer func() {
		if recover() != nil || err != nil {
			_ = tx.Rollback()
		}
	}()

	if userInfo == nil {
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		avatar := utils.GenerateAvatar(uidStr)
		name := randomname.GenerateName()
		male := constant.Male
		user := &model.ScaAuthUser{
			UID:      uidStr,
			Phone:    req.Phone,
			Avatar:   avatar,
			Nickname: name,
			Gender:   male,
		}
		err := tx.ScaAuthUser.Create(user)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		_, err = l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		data, err := HandleLoginJWT(user, l.svcCtx, req.AutoLogin, r, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return response.SuccessWithData(data), nil
	} else {
		data, err := HandleLoginJWT(userInfo, l.svcCtx, req.AutoLogin, r, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		err = tx.Commit()
		if err != nil {
			return nil, err
		}
		return response.SuccessWithData(data), nil
	}
}
