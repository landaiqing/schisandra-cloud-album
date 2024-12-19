package user

import (
	"context"
	"errors"
	"github.com/rbcervilla/redisstore/v9"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/query"
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

	user := l.svcCtx.DB.ScaAuthUser
	var selectedUser query.IScaAuthUserDo

	switch {
	case utils.IsPhone(req.Account):
		selectedUser = user.Where(user.Phone.Eq(req.Account))
	case utils.IsEmail(req.Account):
		selectedUser = user.Where(user.Email.Eq(req.Account))
	case utils.IsUsername(req.Account):
		selectedUser = user.Where(user.Username.Eq(req.Account))
	default:
		return response.ErrorWithI18n(l.ctx, "login.invalidAccount"), nil
	}
	userInfo, err := selectedUser.First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.ErrorWithI18n(l.ctx, "login.userNotRegistered"), nil
		}
		return nil, err
	}

	if !utils.Verify(userInfo.Password, req.Password) {
		return response.ErrorWithI18n(l.ctx, "login.invalidPassword"), nil
	}
	data, err := HandleUserLogin(userInfo, l.svcCtx, req.AutoLogin, r, w, l.ctx)
	if err != nil {
		return nil, err
	}
	// 记录用户登录设备
	if err = GetUserLoginDevice(userInfo.UID, r, l.svcCtx.Ip2Region, l.svcCtx.DB); err != nil {
		return nil, err
	}
	return response.SuccessWithData(data), nil
}

// HandleUserLogin 处理用户登录
func HandleUserLogin(user *model.ScaAuthUser, svcCtx *svc.ServiceContext, autoLogin bool, r *http.Request, w http.ResponseWriter, ctx context.Context) (*types.LoginResponse, error) {
	// 生成jwt token
	accessToken := jwt.GenerateAccessToken(svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: user.UID,
		Type:   constant.JWT_TYPE_ACCESS,
	})
	var days time.Duration
	if autoLogin {
		days = 7 * 24 * time.Hour
	} else {
		days = time.Hour * 24
	}
	refreshToken := jwt.GenerateRefreshToken(svcCtx.Config.Auth.AccessSecret, jwt.RefreshJWTPayload{
		UserID: user.UID,
		Type:   constant.JWT_TYPE_REFRESH,
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
	err = HandlerSession(r, w, user.UID, svcCtx.Session)
	if err != nil {
		return nil, err
	}
	return &data, nil

}

// HandlerSession is a function to set the user_id in the session
func HandlerSession(r *http.Request, w http.ResponseWriter, userID string, redisSession *redisstore.RedisStore) error {
	session, err := redisSession.Get(r, constant.SESSION_KEY)
	if err != nil {
		return err
	}
	session.Values["user_id"] = userID
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return nil
}
