package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"github.com/zeromicro/go-zero/core/logx"
	"xorm.io/xorm"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
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
	uid, ok := session.Values["uid"].(string)
	if !ok {
		return errors.New("user session not found")
	}

	if err = GetUserLoginDevice(uid, r, l.svcCtx.Ip2Region, l.svcCtx.DB, l.ctx); err != nil {
		return err
	}
	return nil
}

// GetUserLoginDevice 获取用户登录设备
func GetUserLoginDevice(userId string, r *http.Request, ip2location *xdb.Searcher, db *xorm.Engine, ctx context.Context) error {
	userAgent := r.Header.Get("User-Agent")
	if userAgent == "" {
		return errors.New("user agent not found")
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

	var device model.ScaAuthUserDevice
	has, err := db.Where("user_id = ? AND ip = ? AND agent = ?", userId, ip, userAgent).Get(&device)
	if err != nil {
		return err
	}

	if !has {
		// 创建新的设备记录
		newDevice := &model.ScaAuthUserDevice{
			UserId:          userId,
			Bot:             isBot,
			Agent:           userAgent,
			Browser:         browser,
			BrowserVersion:  browserVersion,
			EngineName:      engine,
			EngineVersion:   engineVersion,
			Ip:              ip,
			Location:        location,
			OperatingSystem: os,
			Mobile:          mobile,
			Mozilla:         mozilla,
			Platform:        platform,
		}

		affected, err := db.Insert(newDevice)
		if err != nil || affected == 0 {
			return errors.New("create user device failed")
		}
		return nil
	} else {
		// 如果设备存在，执行更新
		device.Bot = isBot
		device.Agent = userAgent
		device.Browser = browser
		device.BrowserVersion = browserVersion
		device.EngineName = engine
		device.EngineVersion = engineVersion
		device.Ip = ip
		device.Location = location
		device.OperatingSystem = os
		device.Mobile = mobile
		device.Mozilla = mozilla
		device.Platform = platform

		affected, err := db.ID(device.Id).Update(&device)
		if err != nil || affected == 0 {
			return errors.New("update user device failed")
		}
		return nil
	}
}
