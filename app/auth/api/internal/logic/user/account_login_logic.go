package user

import (
	"context"
	errors2 "errors"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/api/model/mysql/model"
	"schisandra-album-cloud-microservices/app/auth/api/model/mysql/query"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/errors"
	"schisandra-album-cloud-microservices/common/i18n"
	"schisandra-album-cloud-microservices/common/jwt"
	"schisandra-album-cloud-microservices/common/utils"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
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

func (l *AccountLoginLogic) AccountLogin(r *http.Request, req *types.AccountLoginRequest) (resp *types.LoginResponse, err error) {
	verifyResult := verify.VerifyRotateCaptcha(l.ctx, l.svcCtx.RedisClient, req.Angle, req.Key)
	if !verifyResult {
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "captcha.verificationFailure"))
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
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.invalidAccount"))
	}
	userInfo, err := selectedUser.First()
	if err != nil {
		if errors2.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.notFoundAccount"))
		}
		return nil, err
	}

	if !utils.Verify(userInfo.Password, req.Password) {
		return nil, errors.New(http.StatusInternalServerError, i18n.FormatText(l.ctx, "login.invalidPassword"))
	}
	data, err := HandleLoginJWT(userInfo, l.svcCtx, req.AutoLogin, r, l.ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

// HandleLoginJWT 处理用户登录
func HandleLoginJWT(user *model.ScaAuthUser, svcCtx *svc.ServiceContext, autoLogin bool, r *http.Request, ctx context.Context) (*types.LoginResponse, error) {
	// 获取用户登录设备
	err := GetUserLoginDevice(user.UID, r, svcCtx.Ip2Region, svcCtx.DB)
	if err != nil {
		return nil, err
	}
	// 生成jwt token
	accessToken, expireAt := jwt.GenerateAccessToken(svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: user.UID,
		Type:   constant.JWT_TYPE_ACCESS,
	})
	var days time.Duration
	if autoLogin {
		days = 3 * 24 * time.Hour
	} else {
		days = time.Hour * 24
	}
	refreshToken := jwt.GenerateRefreshToken(svcCtx.Config.Auth.AccessSecret, jwt.RefreshJWTPayload{
		UserID: user.UID,
		Type:   constant.JWT_TYPE_REFRESH,
	}, days)
	data := types.LoginResponse{
		AccessToken: accessToken,
		ExpireAt:    expireAt,
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
		Revoked:      false,
		GeneratedAt:  time.Now().Format(constant.TimeFormat),
		AllowAgent:   r.UserAgent(),
		GeneratedIP:  utils.GetClientIP(r),
		UpdatedAt:    time.Now().Format(constant.TimeFormat),
	}
	err = svcCtx.RedisClient.Set(ctx, constant.UserTokenPrefix+user.UID, redisToken, days).Err()
	if err != nil {
		return nil, err
	}
	return &data, nil
}

// GetUserLoginDevice 获取用户登录设备
func GetUserLoginDevice(userId string, r *http.Request, ip2location *xdb.Searcher, DB *query.Query) error {
	userAgent := r.UserAgent()
	if userAgent == "" {
		return errors2.New("user agent not found")
	}
	ip := utils.GetClientIP(r)
	location, err := ip2location.SearchByStr(ip)
	if err != nil {
		return err
	}
	location = utils.RemoveZeroAndAdjust(location)

	ua := useragent.New(userAgent)
	isBot := ua.Bot()
	browser, browserVersion := ua.Browser()
	os := ua.OS()
	mobile := ua.Mobile()
	mozilla := ua.Mozilla()
	platform := ua.Platform()
	engine, engineVersion := ua.Engine()
	var newIsBot int64 = 0
	var newIsMobile int64 = 0
	if isBot {
		newIsBot = 1
	}
	if mobile {
		newIsMobile = 1
	}
	userDevice := DB.ScaAuthUserDevice
	device, err := userDevice.Where(userDevice.UserID.Eq(userId), userDevice.IP.Eq(ip), userDevice.Agent.Eq(userAgent)).First()
	if err != nil && !errors2.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	newDevice := &model.ScaAuthUserDevice{
		UserID:          userId,
		Bot:             newIsBot,
		Agent:           userAgent,
		Browser:         browser,
		BrowserVersion:  browserVersion,
		EngineName:      engine,
		EngineVersion:   engineVersion,
		IP:              ip,
		Location:        location,
		OperatingSystem: os,
		Mobile:          newIsMobile,
		Mozilla:         mozilla,
		Platform:        platform,
	}
	if device == nil {
		// 创建新的设备记录
		err = DB.ScaAuthUserDevice.Create(newDevice)
		if err != nil {
			return err
		}
		return nil
	} else {
		resultInfo, err := userDevice.Where(userDevice.ID.Eq(device.ID)).Updates(newDevice)
		if err != nil || resultInfo.RowsAffected == 0 {
			return errors2.New("update device failed")
		}
		return nil
	}
}
