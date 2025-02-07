package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math/rand"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sync"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryAllImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryAllImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryAllImageListLogic {
	return &QueryAllImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryAllImageListLogic) QueryAllImageList(req *types.AllImageListRequest) (resp *types.AllImageListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	//  从缓存获取数据
	cacheKey := fmt.Sprintf("%s%s:%s:%s:%t", constant.ImageListPrefix, uid, req.Provider, req.Bucket, req.Sort)
	// 尝试从缓存获取
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse types.AllImageListResponse
		if err := json.Unmarshal([]byte(cachedResult), &cachedResponse); err == nil {
			return &cachedResponse, nil
		}
		logx.Error("Failed to unmarshal cached image list:", err)
		return nil, errors.New("get cached image list failed")
	} else if !errors.Is(err, redis.Nil) {
		logx.Error("Redis error:", err)
		return nil, errors.New("get cached image list failed")
	}

	// 缓存未命中，从数据库中查询
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	// 数据库查询文件信息列表
	var storageInfoQuery query.IScaStorageInfoDo
	if req.Sort {
		storageInfoQuery = storageInfo.Where(storageInfo.UserID.Eq(uid), storageInfo.Provider.Eq(req.Provider), storageInfo.Bucket.Eq(req.Bucket), storageInfo.AlbumID.IsNull()).Order(storageInfo.CreatedAt.Desc())
	} else {
		storageInfoQuery = storageInfo.Where(storageInfo.UserID.Eq(uid), storageInfo.Provider.Eq(req.Provider), storageInfo.Bucket.Eq(req.Bucket)).Order(storageInfo.CreatedAt.Desc())
	}
	storageInfoList, err := storageInfoQuery.Find()
	if err != nil {
		return nil, err
	}
	if len(storageInfoList) == 0 {
		return &types.AllImageListResponse{}, nil
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
		go func(dbFileInfo *model.ScaStorageInfo) {
			defer wg.Done()
			date := dbFileInfo.CreatedAt.Format("2006-01-02")
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
				FilePath:  dbFileInfo.Path,
				URL:       url,
				FileSize:  dbFileInfo.FileSize,
				CreatedAt: dbFileInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			})

			// 重新存储更新后的图像列表
			groupedImages.Store(date, images)
		}(dbFileInfo)
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
	resp = &types.AllImageListResponse{
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
func (l *QueryAllImageListLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(config.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(config.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt secret key failed")
	}
	return &storageConfig.StorageConfig{
		Provider:   config.Type,
		Endpoint:   config.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: config.Bucket,
		Region:     config.Region,
	}, nil
}

// 从缓存或数据库中获取 OSS 配置
func (l *QueryAllImageListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
	dbOssConfig, err := scaOssConfig.Where(scaOssConfig.UserID.Eq(uid), scaOssConfig.Type.Eq(provider)).First()
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
