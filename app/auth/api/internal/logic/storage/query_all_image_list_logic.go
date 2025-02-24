package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gorm.io/gen"
	"math/rand"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sort"
	"sync"
	"time"
)

type QueryAllImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

var WeekdayMap = map[time.Weekday]string{
	time.Sunday:    "日",
	time.Monday:    "一",
	time.Tuesday:   "二",
	time.Wednesday: "三",
	time.Thursday:  "四",
	time.Friday:    "五",
	time.Saturday:  "六",
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
	//  缓存获取数据 v1.0.0
	cacheKey := fmt.Sprintf("%s%s:%s:%s:%s:%s:%t", constant.ImageCachePrefix, uid, "list", req.Provider, req.Bucket, req.Type, req.Sort)
	// 尝试从缓存获取
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err == nil {
		var cachedResponse types.AllImageListResponse
		if err := json.Unmarshal([]byte(cachedResult), &cachedResponse); err == nil {
			return &cachedResponse, nil
		}
		logx.Error("Failed to unmarshal cached image list, deleting invalid cache:", err)
		// 删除无效缓存
		if err := l.svcCtx.RedisClient.Del(l.ctx, cacheKey).Err(); err != nil {
			logx.Error("Failed to delete invalid cache:", err)
		}
	} else if !errors.Is(err, redis.Nil) {
		logx.Error("Redis error:", err)
		return nil, errors.New("get cached image list failed")
	}
	// 缓存未命中，从数据库中查询
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageThumb := l.svcCtx.DB.ScaStorageThumb
	var queryCondition []gen.Condition
	conditions := []gen.Condition{
		storageInfo.UserID.Eq(uid),
		storageInfo.Provider.Eq(req.Provider),
		storageInfo.Bucket.Eq(req.Bucket),
		storageInfo.Type.Neq(constant.ImageTypeShared),
		storageInfo.IsDisplayed.Eq(0),
	}
	queryCondition = append(queryCondition, conditions...)
	if req.Type != "all" {
		queryCondition = append(queryCondition, storageInfo.Type.Eq(req.Type))
	}
	if req.Sort {
		queryCondition = append(queryCondition, storageInfo.AlbumID.Eq(0))
	}
	var storageInfoList []types.FileInfoResult
	err = storageInfo.Select(
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
		Order(storageInfo.CreatedAt.Desc()).Scan(&storageInfoList)
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
	g, ctx := errgroup.WithContext(l.ctx)
	sem := semaphore.NewWeighted(10) // 限制并发数为 10
	groupedImages := sync.Map{}

	for _, dbFileInfo := range storageInfoList {
		dbFileInfo := dbFileInfo // 创建局部变量以避免闭包问题
		if err := sem.Acquire(ctx, 1); err != nil {
			logx.Error("Failed to acquire semaphore:", err)
			continue
		}
		g.Go(func() error {
			defer sem.Release(1)
			weekday := WeekdayMap[dbFileInfo.CreatedAt.Weekday()]
			date := dbFileInfo.CreatedAt.Format("2006年1月2日 星期" + weekday)
			reqParams := make(url.Values)
			presignedUrl, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ThumbnailBucketName, dbFileInfo.ThumbPath, time.Hour*24*7, reqParams)
			if err != nil {
				logx.Error(err)
				return err
			}
			url, err := service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.Path, time.Hour*24*7)
			if err != nil {
				logx.Error(err)
				return err
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
			return nil
		})
	}
	// 等待所有 goroutine 完成
	if err = g.Wait(); err != nil {
		return nil, err
	}
	var imageList []types.AllImageDetail
	groupedImages.Range(func(key, value interface{}) bool {
		imageList = append(imageList, types.AllImageDetail{
			Date: key.(string),
			List: value.([]types.ImageMeta),
		})
		return true
	})
	sort.Slice(imageList, func(i, j int) bool {
		if len(imageList[i].List) == 0 || len(imageList[j].List) == 0 {
			return false // 空列表不参与排序
		}
		createdAtI, _ := time.Parse("2006-01-02 15:04:05", imageList[i].List[0].CreatedAt)
		createdAtJ, _ := time.Parse("2006-01-02 15:04:05", imageList[j].List[0].CreatedAt)
		return createdAtI.After(createdAtJ) // 降序排序
	})
	resp = &types.AllImageListResponse{
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
		Provider:   config.Provider,
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
