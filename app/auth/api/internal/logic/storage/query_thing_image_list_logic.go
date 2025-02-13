package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryThingImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryThingImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryThingImageListLogic {
	return &QueryThingImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryThingImageListLogic) QueryThingImageList(req *types.ThingListRequest) (resp *types.ThingListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageInfos, err := storageInfo.Select(
		storageInfo.ID,
		storageInfo.Category,
		storageInfo.Tag,
		storageInfo.Path,
		storageInfo.CreatedAt).
		Where(storageInfo.UserID.Eq(uid),
			storageInfo.Provider.Eq(req.Provider),
			storageInfo.Bucket.Eq(req.Bucket),
			storageInfo.Category.IsNotNull(),
			storageInfo.Tag.IsNotNull(),
			storageInfo.Category.Length().Gt(0),
			storageInfo.Tag.Length().Gte(0)).
		Order(storageInfo.CreatedAt.Desc()).
		Find()
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

	categoryMap := sync.Map{}
	tagCountMap := sync.Map{}
	tagCoverMap := sync.Map{} // 用于存储每个 Tag 的封面图片路径

	for _, info := range storageInfos {
		tagKey := info.Category + "::" + info.Tag
		if _, exists := tagCountMap.Load(tagKey); !exists {
			tagCountMap.Store(tagKey, int64(0))
			categoryEntry, _ := categoryMap.LoadOrStore(info.Category, &sync.Map{})
			tagMap := categoryEntry.(*sync.Map)
			tagMap.Store(info.Tag, types.ThingMeta{
				TagName:   info.Tag,
				CreatedAt: info.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		tagCount, _ := tagCountMap.Load(tagKey)
		tagCountMap.Store(tagKey, tagCount.(int64)+1)

		// 为每个 Tag 存储封面图片路径
		if _, exists := tagCoverMap.Load(tagKey); !exists {
			// 使用服务生成预签名 URL
			coverImageURL, err := service.PresignedURL(l.ctx, req.Bucket, info.Path, 7*24*time.Hour)
			if err == nil {
				tagCoverMap.Store(tagKey, coverImageURL)
			}
		}
	}

	var thingListData []types.ThingListData
	categoryMap.Range(func(category, tagData interface{}) bool {
		var metas []types.ThingMeta
		tagData.(*sync.Map).Range(func(tag, item interface{}) bool {
			tagKey := category.(string) + "::" + tag.(string)
			tagCount, _ := tagCountMap.Load(tagKey)
			meta := item.(types.ThingMeta)
			meta.TagCount = tagCount.(int64)

			// 获取封面图片 URL
			if coverImageURL, ok := tagCoverMap.Load(tagKey); ok {
				meta.CoverImage = coverImageURL.(string)
			}
			metas = append(metas, meta)
			return true
		})
		thingListData = append(thingListData, types.ThingListData{
			Category: category.(string),
			List:     metas,
		})
		return true
	})

	return &types.ThingListResponse{Records: thingListData}, nil
}

// 提取解密操作为函数
func (l *QueryThingImageListLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *QueryThingImageListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
