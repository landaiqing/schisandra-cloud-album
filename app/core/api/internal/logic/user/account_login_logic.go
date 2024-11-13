package user

import (
	"context"
	"net/http"
	"time"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"github.com/zeromicro/go-zero/core/logc"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/captcha/verify"
	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/i18n"
	"schisandra-album-cloud-microservices/app/core/api/common/jwt"
	"schisandra-album-cloud-microservices/app/core/api/common/response"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuser"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuserdevice"
	types3 "schisandra-album-cloud-microservices/app/core/api/repository/redis_session/types"
	types2 "schisandra-album-cloud-microservices/app/core/api/repository/redisx/types"
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
		return response.ErrorWithMessage(i18n.FormatText(l.ctx, "captcha.verificationFailure", "验证失败！")), nil
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
		return response.ErrorWithMessage(i18n.FormatText(l.ctx, "login.invalidAccount", "无效账号！")), nil
	}

	user, err = query.First(l.ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return response.ErrorWithMessage(i18n.FormatText(l.ctx, "login.notFoundAccount", "无效账号！")), nil
		}
		return nil, err
	}

	if !utils.Verify(user.Password, req.Password) {
		return response.ErrorWithMessage(i18n.FormatText(l.ctx, "login.invalidPassword", "密码错误！")), nil
	}
	data, result := HandleUserLogin(user, l, req.AutoLogin, r, w, l.svcCtx.Ip2Region, l.svcCtx.MySQLClient)
	if !result {
		return response.ErrorWithMessage(i18n.FormatText(l.ctx, "login.loginFailed", "登录失败！")), nil
	}
	return response.Success(data), nil
}

// HandleUserLogin 处理用户登录
func HandleUserLogin(user *ent.ScaAuthUser, l *AccountLoginLogic, autoLogin bool, r *http.Request, w http.ResponseWriter, ip2location *xdb.Searcher, entClient *ent.Client) (*types.LoginResponse, bool) {
	// 生成jwt token
	accessToken := jwt.GenerateAccessToken(l.svcCtx.Config.Auth.AccessSecret, jwt.AccessJWTPayload{
		UserID: user.UID,
	})
	var days time.Duration
	if autoLogin {
		days = 7 * 24 * time.Hour
	} else {
		days = time.Hour * 24
	}
	refreshToken := jwt.GenerateRefreshToken(l.svcCtx.Config.Auth.AccessSecret, jwt.RefreshJWTPayload{
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

	redisToken := types2.RedisToken{
		AccessToken: accessToken,
		UID:         user.UID,
	}
	err := l.svcCtx.RedisClient.SetEx(l.ctx, constant.UserTokenPrefix+user.UID, redisToken, days).Err()
	if err != nil {
		logc.Error(l.ctx, err)
		return nil, false
	}
	sessionData := types3.SessionData{
		RefreshToken: refreshToken,
		UID:          user.UID,
	}
	session, err := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		logc.Error(l.ctx, err)
		return nil, false
	}
	session.Values[constant.SESSION_KEY] = sessionData
	if err = session.Save(r, w); err != nil {
		logc.Error(l.ctx, err)
		return nil, false
	}
	// 记录用户登录设备
	if !getUserLoginDevice(user.UID, r, ip2location, entClient, l.ctx) {
		return nil, false
	}
	return &data, true

}

// getUserLoginDevice 获取用户登录设备
func getUserLoginDevice(userId string, r *http.Request, ip2location *xdb.Searcher, entClient *ent.Client, ctx context.Context) bool {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return false
	}
	ip := utils.GetClientIP(r)
	location, err := ip2location.SearchByStr(ip)
	if err != nil {
		return false
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

	device, err := entClient.ScaAuthUserDevice.Query().
		Where(scaauthuserdevice.UserID(userId), scaauthuserdevice.IP(ip), scaauthuserdevice.Agent(userAgent)).
		Only(ctx)

	// 如果有错误，表示设备不存在，执行插入
	if ent.IsNotFound(err) {
		// 创建新的设备记录
		entClient.ScaAuthUserDevice.Create().
			SetBot(isBot).
			SetAgent(userAgent).
			SetBrowser(browser).
			SetBrowserVersion(browserVersion).
			SetEngineName(engine).
			SetEngineVersion(engineVersion).
			SetIP(ip).
			SetLocation(location).
			SetOperatingSystem(os).
			SetMobile(mobile).
			SetMozilla(mozilla).
			SetPlatform(platform).
			SaveX(ctx)
	} else if err == nil {
		// 如果设备存在，执行更新
		device.Update().
			SetBot(isBot).
			SetAgent(userAgent).
			SetBrowser(browser).
			SetBrowserVersion(browserVersion).
			SetEngineName(engine).
			SetEngineVersion(engineVersion).
			SetIP(ip).
			SetLocation(location).
			SetOperatingSystem(os).
			SetMobile(mobile).
			SetMozilla(mozilla).
			SetPlatform(platform).
			SaveX(ctx)
	} else {
		logc.Error(ctx, err)
		return false
	}
	return true
}
