package user

import (
	"context"
	"errors"
	"github.com/yitter/idgenerator-go/idgen"
	"gorm.io/gorm"
	"net/http"
	model2 "schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	errors2 "schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/random_name"
	"schisandra-album-cloud-microservices/common/utils"
	"strconv"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type WechatOffiaccountLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewWechatOffiaccountLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *WechatOffiaccountLoginLogic {
	return &WechatOffiaccountLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *WechatOffiaccountLoginLogic) WechatOffiaccountLogin(r *http.Request, req *types.WechatOffiaccountLoginRequest) (resp *types.LoginResponse, err error) {
	decryptedClientId, err := encrypt.Decrypt(req.ClientId, l.svcCtx.Config.Encrypt.Key, l.svcCtx.Config.Encrypt.IV)
	if err != nil {
		return nil, err
	}
	clientId := l.svcCtx.RedisClient.Get(r.Context(), constant.UserClientPrefix+decryptedClientId).Val()
	if clientId == "" {
		return nil, errors2.New(http.StatusUnauthorized, i18n.FormatText(l.ctx, "login.loginFailed"))
	}
	Openid, err := encrypt.Decrypt(req.Openid, l.svcCtx.Config.Encrypt.Key, l.svcCtx.Config.Encrypt.IV)
	if err != nil {
		return nil, err
	}
	tx := l.svcCtx.DB.Begin()
	userSocial := l.svcCtx.DB.ScaAuthUserSocial
	socialUser, err := tx.ScaAuthUserSocial.Where(userSocial.OpenID.Eq(Openid), userSocial.Source.Eq(constant.OAuthSourceWechat)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if socialUser == nil {
		// 创建用户
		uid := idgen.NextId()
		uidStr := strconv.FormatInt(uid, 10)
		avatar := utils.GenerateAvatar(uidStr)
		name := randomname.GenerateName()

		addUser := &model2.ScaAuthUser{
			UID:      uidStr,
			Avatar:   avatar,
			Username: Openid,
			Nickname: name,
			Gender:   constant.Male,
		}
		err = tx.ScaAuthUser.Create(addUser)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		newSocialUser := &model2.ScaAuthUserSocial{
			UserID: uidStr,
			OpenID: Openid,
			Source: constant.OAuthSourceWechat,
		}
		err = tx.ScaAuthUserSocial.Create(newSocialUser)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		if res, err := l.svcCtx.CasbinEnforcer.AddRoleForUser(uidStr, constant.User); !res || err != nil {
			_ = tx.Rollback()
			return nil, err
		}

		data, err := HandleLoginJWT(addUser, l.svcCtx, true, r, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return data, nil
	} else {
		authUser := l.svcCtx.DB.ScaAuthUser

		authUserInfo, err := tx.ScaAuthUser.Where(authUser.UID.Eq(socialUser.UserID)).First()
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			_ = tx.Rollback()
			return nil, err
		}
		data, err := HandleLoginJWT(authUserInfo, l.svcCtx, true, r, l.ctx)
		if err != nil {
			_ = tx.Rollback()
			return nil, err
		}
		if err = tx.Commit(); err != nil {
			return nil, err
		}
		return data, nil

	}
}
