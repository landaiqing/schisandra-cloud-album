package user

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"xorm.io/xorm"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
)

type AccountLoginLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAccountLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AccountLoginLogic {
	return &AccountLoginLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AccountLoginLogic) AccountLogin(w http.ResponseWriter, r *http.Request, req *types.AccountLoginRequest) (resp *types.Response, err error) {
	verifyResult := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !verifyResult {
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure"), nil
	}
	var user model.ScaAuthUser
	var query *xorm.Session

	switch {
	case utils.IsPhone(req.Account):
		query = l.svcCtx.DB.Where("phone = ? AND deleted = ?", req.Account, 0)
	case utils.IsEmail(req.Account):
		query = l.svcCtx.DB.Where("email = ? AND deleted = ?", req.Account, 0)
	case utils.IsUsername(req.Account):
		query = l.svcCtx.DB.Where("username = ? AND deleted = ?", req.Account, 0)
	default:
		return response.ErrorWithI18n(l.ctx, "login.invalidAccount"), nil
	}
	has, err := query.Get(&user)
	if err != nil {
		return nil, err
	}
	if !has {
		return response.ErrorWithI18n(l.ctx, "login.userNotRegistered"), nil
	}

	if !utils.Verify(user.Password, req.Password) {
		return response.ErrorWithI18n(l.ctx, "login.invalidPassword"), nil
	}
	data, err := HandleUserLogin(user, l.svcCtx, req.AutoLogin, r, w, l.ctx)
	if err != nil {
		return nil, err
	}
	// 记录用户登录设备
	if err = GetUserLoginDevice(user.UID, r, l.svcCtx.Ip2Region, l.svcCtx.DB, l.ctx); err != nil {
		return nil, err
	}
	return response.SuccessWithData(data), nil
}

// HandleUserLogin 处理用户登录
func HandleUserLogin(user model.ScaAuthUser, svcCtx *svc.ServiceContext, autoLogin bool, r *http.Request, w http.ResponseWriter, ctx context.Context) (*types.LoginResponse, error) {
	// 生成jwt token
	accessToken := jwt.GenerateAccessToken(svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: user.UID,
	})
	var days time.Duration
	if autoLogin {
		days = 7 * 24 * time.Hour
	} else {
		days = time.Hour * 24
	}
	refreshToken := jwt.GenerateRefreshToken(svcCtx.Config.Auth.AccessSecret, jwt.RefreshJWTPayload{
		UserID: user.UID,
	}, days)
	data := types.LoginResponse{
		AccessToken: accessToken,
		UID:         user.UID,
		Username:    user.Username,
		Nickname:    user.Nickname,
		Avatar:      user.Avatar,
		Status:      user.Status,
	}

	redisToken := types.RedisToken{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		UID:          user.UID,
	}
	err := svcCtx.RedisClient.Set(ctx, constant.UserTokenPrefix+user.UID, redisToken, days).Err()
	if err != nil {
		return nil, err
	}
	session, err := svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		return nil, err
	}
	session.Values["refresh_token"] = refreshToken
	session.Values["uid"] = user.UID
	err = session.Save(r, w)
	if err != nil {
		return nil, err
	}
	return &data, nil

}
