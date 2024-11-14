package user

import (
	"context"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuser"
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
		return response.ErrorWithI18n(l.ctx, "captcha.verificationFailure", "验证失败！"), nil
	}
	var user *ent.ScaAuthUser
	var query *ent.ScaAuthUserQuery

	switch {
	case utils.IsPhone(req.Account):
		query = l.svcCtx.MySQLClient.ScaAuthUser.Query().Where(scaauthuser.PhoneEQ(req.Account), scaauthuser.DeletedEQ(0))
	case utils.IsEmail(req.Account):
		query = l.svcCtx.MySQLClient.ScaAuthUser.Query().Where(scaauthuser.EmailEQ(req.Account), scaauthuser.DeletedEQ(0))
	case utils.IsUsername(req.Account):
		query = l.svcCtx.MySQLClient.ScaAuthUser.Query().Where(scaauthuser.UsernameEQ(req.Account), scaauthuser.DeletedEQ(0))
	default:
		return response.ErrorWithI18n(l.ctx, "login.invalidAccount", "无效账号！"), nil
	}

	user, err = query.First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return response.ErrorWithI18n(l.ctx, "login.userNotRegistered", "用户未注册！"), nil
		}
		return nil, err
	}

	if !utils.Verify(user.Password, req.Password) {
		return response.ErrorWithI18n(l.ctx, "login.invalidPassword", "密码错误！"), nil
	}
	data, result := HandleUserLogin(user, l.svcCtx, req.AutoLogin, r, w, l.ctx)
	if !result {
		return response.ErrorWithI18n(l.ctx, "login.loginFailed", "登录失败！"), nil
	}
	return response.SuccessWithData(data), nil
}

// HandleUserLogin 处理用户登录
func HandleUserLogin(user *ent.ScaAuthUser, svcCtx *svc.ServiceContext, autoLogin bool, r *http.Request, w http.ResponseWriter, ctx context.Context) (*types.LoginResponse, bool) {
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
		AccessToken: accessToken,
		UID:         user.UID,
	}
	err := svcCtx.RedisClient.Set(ctx, constant.UserTokenPrefix+user.UID, redisToken, days).Err()
	if err != nil {
		logc.Error(ctx, err)
		return nil, false
	}
	sessionData := types.SessionData{
		RefreshToken: refreshToken,
		UID:          user.UID,
	}
	session, err := svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		logc.Error(ctx, err)
		return nil, false
	}
	session.Values[constant.SESSION_KEY] = sessionData
	err = session.Save(r, w)
	if err != nil {
		return nil, false
	}
	// 记录用户登录设备
	if !GetUserLoginDevice(user.UID, r, svcCtx.Ip2Region, svcCtx.MySQLClient, ctx) {
		return nil, false
	}
	return &data, true

}
