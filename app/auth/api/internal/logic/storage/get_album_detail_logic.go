package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"gorm.io/gen"
	"math/rand"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sort"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAlbumDetailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAlbumDetailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAlbumDetailLogic {
	return &GetAlbumDetailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAlbumDetailLogic) GetAlbumDetail(req *types.AlbumDetailListRequest) (resp *types.AlbumDetailListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	//  缓存获取数据 v1.0.0
	cacheKey := fmt.Sprintf("%s%s:%s:%s:%s:%v", constant.ImageCachePrefix, uid, "album", req.Provider, req.Bucket, req.ID)
	// 尝试从缓存获取
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse types.AlbumDetailListResponse
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
	var queryCondition []gen.Condition
	queryCondition = append(queryCondition, storageInfo.UserID.Eq(uid))
	queryCondition = append(queryCondition, storageInfo.AlbumID.Eq(req.ID))
	// 类型筛选 1 是分享类型
	if req.Type != constant.AlbumTypeShared {
		queryCondition = append(queryCondition, storageInfo.Provider.Eq(req.Provider))
		queryCondition = append(queryCondition, storageInfo.Bucket.Eq(req.Bucket))
	}

	storageInfoQuery = storageInfo.Select(
		storageInfo.ID,
		storageInfo.FileName,
		storageInfo.CreatedAt,
		storageThumb.ThumbPath,
		storageInfo.Path,
		storageThumb.ThumbW,
		storageThumb.ThumbH,
		storageThumb.ThumbSize).
		LeftJoin(storageThumb, storageInfo.ID.EqCol(storageThumb.InfoID)).
		Where(queryCondition...).
		Order(storageInfo.CreatedAt.Desc())
	err = storageInfoQuery.Scan(&storageInfoList)
	if err != nil {
		return nil, err
	}
	if len(storageInfoList) == 0 {
		return &types.AlbumDetailListResponse{}, nil
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
	reqParams := make(url.Values)
	for _, dbFileInfo := range storageInfoList {
		wg.Add(1)
		go func(dbFileInfo *types.FileInfoResult) {
			defer wg.Done()
			weekday := WeekdayMap[dbFileInfo.CreatedAt.Weekday()]
			date := dbFileInfo.CreatedAt.Format("2006年1月2日 星期" + weekday)

			var thumbnailUrl string
			var originalUrl string

			if req.Type == constant.AlbumTypeShared {
				minioOriginalUrl, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ShareImagesBucketName, dbFileInfo.Path, 30*time.Minute, reqParams)
				originalUrl = minioOriginalUrl.String()
				if err != nil {
					return
				}
				minioThumbnailUrl, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ThumbnailBucketName, dbFileInfo.ThumbPath, 30*time.Minute, reqParams)
				thumbnailUrl = minioThumbnailUrl.String()
				if err != nil {
					return
				}
			} else {
				thumbnailUrl, err = service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.ThumbPath, time.Minute*30)
				if err != nil {
					logx.Error(err)
					return
				}
				originalUrl, err = service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.Path, time.Minute*30)
				if err != nil {
					logx.Error(err)
					return
				}
			}

			// 使用 Load 或 Store 确保原子操作
			value, _ := groupedImages.LoadOrStore(date, []types.ImageMeta{})
			images := value.([]types.ImageMeta)

			images = append(images, types.ImageMeta{
				ID:        dbFileInfo.ID,
				FileName:  dbFileInfo.FileName,
				Thumbnail: thumbnailUrl,
				URL:       originalUrl,
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
	// 按日期排序，最新的在最上面
	sort.Slice(imageList, func(i, j int) bool {
		dateI, _ := time.Parse("2006年1月2日 星期一", imageList[i].Date)
		dateJ, _ := time.Parse("2006年1月2日 星期一", imageList[j].Date)
		return dateI.After(dateJ)
	})
	resp = &types.AlbumDetailListResponse{
		Records: imageList,
	}

	// 缓存结果
	if data, err := json.Marshal(resp); err == nil {
		expireTime := 5*time.Minute + time.Duration(rand.Intn(300))*time.Second
		if err := l.svcCtx.RedisClient.Set(l.ctx, cacheKey, data, expireTime).Err(); err != nil {
			logx.Error("Failed to cache image list:", err)
		}
	} else {
		logx.Error("Failed to marshal image list for caching:", err)
	}

	return resp, nil
}

// 提取解密操作为函数
func (l *GetAlbumDetailLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *GetAlbumDetailLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
