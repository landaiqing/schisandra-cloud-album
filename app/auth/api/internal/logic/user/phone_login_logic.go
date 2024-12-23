package user

import (
	"context"
	"errors"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/model/mysql/model"
	constant2 "schisandra-album-cloud-microservices/common/constant"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/random_name"
	utils2 "schisandra-album-cloud-microservices/common/utils"
	"strconv"

	"github.com/yitter/idgenerator-go/idgen"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
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

func (l *PhoneLoginLogic) PhoneLogin(r *http.Request, req *types.PhoneLoginRequest) (resp *types.LoginResponse, err error) {
	if !utils2.IsPhone(req.Phone) {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
	}
	code := l.svcCtx.RedisClient.Get(l.ctx, constant2.UserSmsRedisPrefix+req.Phone).Val()
	if code == "" {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.captchaExpired"))
	}
	if req.Captcha != code {
		return nil, errors2.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.captchaError"))
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
		avatar := utils2.GenerateAvatar(uidStr)
		name := randomname.GenerateName()
		male := constant2.Male
		user := &model.ScaAuthUser{
			UID:      uidStr,
			Phone:    req.Phone,
			Avatar:   avatar,
			Nickname: name,
			Gender:   male,
		}
		err = tx.ScaAuthUser.Create(user)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		_, err = l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant2.User)
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
		return data, nil
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
		return data, nil
	}
}
