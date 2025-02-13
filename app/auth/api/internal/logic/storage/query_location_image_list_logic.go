package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryLocationImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLocationImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLocationImageListLogic {
	return &QueryLocationImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryLocationImageListLogic) QueryLocationImageList(req *types.LocationListRequest) (resp *types.LocationListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageLocation := l.svcCtx.DB.ScaStorageLocation

	locations, err := storageLocation.Select(
		storageLocation.ID,
		storageLocation.Country,
		storageLocation.City,
		storageLocation.Province,
		storageLocation.CoverImage,
		storageLocation.Total).Where(storageLocation.UserID.Eq(uid),
		storageLocation.Provider.Eq(req.Provider),
		storageLocation.Bucket.Eq(req.Bucket)).
		Order(storageLocation.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}

	// 加载用户oss配置信息
	cacheOssConfigKey := constant.UserOssConfigPrefix + uid + ":" + req.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheOssConfigKey, uid, req.Provider)
	if err != nil {
		return nil, err
	}

	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return nil, errors.New("get storage failed")
	}
	locationMap := make(map[string][]types.LocationMeta)

	for _, loc := range locations {
		var locationKey string
		if loc.Province == "" {
			locationKey = loc.Country
		} else {
			locationKey = fmt.Sprintf("%s %s", loc.Country, loc.Province)
		}

		city := loc.City
		if city == "" {
			city = loc.Country
		}
		url, err := service.PresignedURL(l.ctx, req.Bucket, loc.CoverImage, 7*24*time.Hour)
		if err != nil {
			return nil, errors.New("get presigned url failed")
		}
		locationMeta := types.LocationMeta{
			ID:         loc.ID,
			City:       city,
			Total:      loc.Total,
			CoverImage: url,
		}
		locationMap[locationKey] = append(locationMap[locationKey], locationMeta)
	}

	var locationListData []types.LocationListData

	for location, list := range locationMap {
		locationListData = append(locationListData, types.LocationListData{
			Location: location,
			List:     list,
		})
	}

	return &types.LocationListResponse{Records: locationListData}, nil
}

// 提取解密操作为函数
func (l *QueryLocationImageListLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(config.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(config.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt secret key failed")
	}
	return &storageConfig.StorageConfig{
		Provider:   config.Provider,
		Endpoint:   config.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: config.Bucket,
		Region:     config.Region,
	}, nil
}

// 从缓存或数据库中获取 OSS 配置
func (l *QueryLocationImageListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
	result, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.New("get oss config failed")
	}

	var ossConfig *storageConfig.StorageConfig
	if result != "" {
		var redisOssConfig model.ScaStorageConfig
		if err = json.Unmarshal([]byte(result), &redisOssConfig); err != nil {
			return nil, errors.New("unmarshal oss config failed")
		}
		return l.decryptConfig(&redisOssConfig)
	}

	// 缓存未命中，从数据库中加载
	scaOssConfig := l.svcCtx.DB.ScaStorageConfig
	dbOssConfig, err := scaOssConfig.Where(scaOssConfig.UserID.Eq(uid), scaOssConfig.Provider.Eq(provider)).First()
	if err != nil {
		return nil, err
	}

	// 缓存数据库配置
	ossConfig, err = l.decryptConfig(dbOssConfig)
	if err != nil {
		return nil, err
	}
	marshalData, err := json.Marshal(dbOssConfig)
	if err != nil {
		return nil, errors.New("marshal oss config failed")
	}
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, marshalData, 0).Err()
	if err != nil {
		return nil, errors.New("set oss config failed")
	}

	return ossConfig, nil
}
