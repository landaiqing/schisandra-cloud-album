package user

import (
	"context"
	"errors"
	"net/http"

	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mssola/useragent"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"

	"schisandra-album-cloud-microservices/app/core/api/common/constant"
	"schisandra-album-cloud-microservices/app/core/api/common/utils"
	"schisandra-album-cloud-microservices/app/core/api/internal/svc"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/model"
	"schisandra-album-cloud-microservices/app/core/api/repository/mysql/query"
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

	if err = GetUserLoginDevice(uid, r, l.svcCtx.Ip2Region, l.svcCtx.DB); err != nil {
		return err
	}
	return nil
}

// GetUserLoginDevice 获取用户登录设备
func GetUserLoginDevice(userId string, r *http.Request, ip2location *xdb.Searcher, DB *query.Query) error {
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
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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
			return errors.New("update device failed")
		}
		return nil
	}
}
