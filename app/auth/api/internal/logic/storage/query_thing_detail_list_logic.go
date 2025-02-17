package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryThingDetailListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryThingDetailListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryThingDetailListLogic {
	return &QueryThingDetailListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryThingDetailListLogic) QueryThingDetailList(req *types.ThingDetailListRequest) (resp *types.ThingDetailListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	//  缓存获取数据 v1.0.0
	cacheKey := fmt.Sprintf("%s%s:%s:%s:%v", constant.ImageListPrefix, uid, req.Provider, req.Bucket, req.TagName)
	// 尝试从缓存获取
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse types.ThingDetailListResponse
		if err := json.Unmarshal([]byte(cachedResult), &cachedResponse); err == nil {
			return &cachedResponse, nil
		}
		logx.Error("Failed to unmarshal cached image list:", err)
		return nil, errors.New("get cached image list failed")
	} else if !errors.Is(err, redis.Nil) {
		logx.Error("Redis error:", err)
		return nil, errors.New("get cached image list failed")
	}

	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageThumb := l.svcCtx.DB.ScaStorageThumb
	// 数据库查询文件信息列表
	var storageInfoQuery query.IScaStorageInfoDo
	var storageInfoList []types.FileInfoResult

	storageInfoQuery = storageInfo.Select(
		storageInfo.ID,
		storageInfo.FileName,
		storageInfo.CreatedAt,
		storageThumb.ThumbPath,
		storageInfo.Path,
		storageThumb.ThumbW,
		storageThumb.ThumbH,
		storageThumb.ThumbSize).
		LeftJoin(storageThumb, storageInfo.ThumbID.EqCol(storageThumb.ID)).
		Where(
			storageInfo.UserID.Eq(uid),
			storageInfo.Provider.Eq(req.Provider),
			storageInfo.Bucket.Eq(req.Bucket),
			storageInfo.Tag.Eq(req.TagName)).
		Order(storageInfo.CreatedAt.Desc())
	err = storageInfoQuery.Scan(&storageInfoList)
	if err != nil {
		return nil, err
	}
	if len(storageInfoList) == 0 {
		return &types.ThingDetailListResponse{}, nil
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

	// 按日期进行分组
	var wg sync.WaitGroup
	groupedImages := sync.Map{}

	for _, dbFileInfo := range storageInfoList {
		wg.Add(1)
		go func(dbFileInfo *types.FileInfoResult) {
			defer wg.Done()
			weekday := WeekdayMap[dbFileInfo.CreatedAt.Weekday()]
			date := dbFileInfo.CreatedAt.Format("2006年1月2日 星期" + weekday)
			reqParams := make(url.Values)
			presignedUrl, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ThumbnailBucketName, dbFileInfo.ThumbPath, time.Hour*24*7, reqParams)
			if err != nil {
				logx.Error(err)
				return
			}
			url, err := service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.Path, time.Hour*24*7)
			if err != nil {
				logx.Error(err)
				return
			}
			// 使用 Load 或 Store 确保原子操作
			value, _ := groupedImages.LoadOrStore(date, []types.ImageMeta{})
			images := value.([]types.ImageMeta)

			images = append(images, types.ImageMeta{
				ID:        dbFileInfo.ID,
				FileName:  dbFileInfo.FileName,
				Thumbnail: presignedUrl.String(),
				URL:       url,
				Width:     dbFileInfo.ThumbW,
				Height:    dbFileInfo.ThumbH,
				CreatedAt: dbFileInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			})

			// 重新存储更新后的图像列表
			groupedImages.Store(date, images)
		}(&dbFileInfo)
	}
	wg.Wait()
	var imageList []types.AllImageDetail
	groupedImages.Range(func(key, value interface{}) bool {
		imageList = append(imageList, types.AllImageDetail{
			Date: key.(string),
			List: value.([]types.ImageMeta),
		})
		return true
	})
	resp = &types.ThingDetailListResponse{
		Records: imageList,
	}

	// 缓存结果
	if data, err := json.Marshal(resp); err == nil {
		expireTime := 7*24*time.Hour - time.Duration(rand.Intn(60))*time.Minute
		if err := l.svcCtx.RedisClient.Set(l.ctx, cacheKey, data, expireTime).Err(); err != nil {
			logx.Error("Failed to cache image list:", err)
		}
	} else {
		logx.Error("Failed to marshal image list for caching:", err)
	}

	return resp, nil
}

// 提取解密操作为函数
func (l *QueryThingDetailListLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *QueryThingDetailListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
