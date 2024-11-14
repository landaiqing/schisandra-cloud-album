package user

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"github.com/zeromicro/go-zero/core/logx"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/internal/types"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/ent/scaauthuserdevice"
)

type GetUserDeviceLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserDeviceLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserDeviceLogic {
	return &GetUserDeviceLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserDeviceLogic) GetUserDevice(r *http.Request) error {
	session, err := l.svcCtx.Session.Get(r, constant.SESSION_KEY)
	if err != nil {
		return err
	}
	sessionData, ok := session.Values[constant.SESSION_KEY]
	if !ok {
		return errors.New("User not found or device not found")
	}
	var data types.SessionData
	err = json.Unmarshal(sessionData.([]byte), &data)
	if err != nil {
		return err
	}

	res := GetUserLoginDevice(data.UID, r, l.svcCtx.Ip2Region, l.svcCtx.MySQLClient, l.ctx)
	if !res {
		return errors.New("User not found or device not found")
	}
	return nil
}

// GetUserLoginDevice 获取用户登录设备
func GetUserLoginDevice(userId string, r *http.Request, ip2location *xdb.Searcher, entClient *ent.Client, ctx context.Context) bool {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return false
	}
	ip := utils.GetClientIP(r)
	location, err := ip2location.SearchByStr(ip)
	if err != nil {
		logx.Error(err)
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
		err = entClient.ScaAuthUserDevice.Create().
			SetUserID(userId).
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
			Exec(ctx)
		if err != nil {
			logx.Error(err)
			return false
		}
		return true
	} else if err == nil {
		// 如果设备存在，执行更新
		err = device.Update().
			SetUserID(userId).
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
			Exec(ctx)
		if err != nil {
			logx.Error(err)
			return false
		}
		return true
	} else {
		logx.Error(err)
		return false
	}
}
