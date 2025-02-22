package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"math"
	"math/rand"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetBucketCapacityLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetBucketCapacityLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetBucketCapacityLogic {
	return &GetBucketCapacityLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetBucketCapacityLogic) GetBucketCapacity(req *types.BucketCapacityRequest) (resp *types.BucketCapacityResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	// 设计缓存键
	cacheKey := fmt.Sprintf("%s%s:%s:%s", constant.BucketCapacityCachePrefix, uid, req.Provider, req.Bucket)
	// 尝试从缓存中获取容量信息
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		logx.Errorf("get bucket capacity from cache failed: %v", err)
		return nil, err
	}

	// 如果缓存存在，直接返回缓存结果
	if cachedResult != "" {
		// 如果是空值缓存（防缓存穿透），返回空结果
		if cachedResult == "{}" {
			return &types.BucketCapacityResponse{}, nil
		}

		err = json.Unmarshal([]byte(cachedResult), &resp)
		if err != nil {
			return nil, errors.New("unmarshal cached result failed")
		}
		return resp, nil
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
	bucketStat, err := service.GetBucketStat(l.ctx, ossConfig.BucketName)
	if err != nil {
		// 如果 OSS 接口调用失败，设置空值缓存（防缓存穿透）
		emptyData := "{}"
		emptyCacheExpire := 5 * time.Minute // 空值缓存过期时间
		if err := l.svcCtx.RedisClient.Set(l.ctx, cacheKey, emptyData, emptyCacheExpire).Err(); err != nil {
			logx.Errorf("set empty cache failed: %v", err)
		}
		return nil, errors.New("get bucket stat failed")
	}
	scaStorageConfig := l.svcCtx.DB.ScaStorageConfig
	capacity, err := scaStorageConfig.Select(scaStorageConfig.Capacity).
		Where(scaStorageConfig.UserID.Eq(uid), scaStorageConfig.Provider.Eq(req.Provider), scaStorageConfig.Bucket.Eq(req.Bucket)).First()
	if err != nil {
		return nil, errors.New("get storage config failed")
	}

	// 总容量（单位：GB）
	totalCapacityGB := capacity.Capacity

	// 已用容量（单位：字节转换为 GB）
	const bytesToGB = 1024 * 1024 * 1024
	usedCapacityGB := float64(bucketStat.StandardStorage) / bytesToGB

	// 计算百分比
	percentage := calculatePercentage(usedCapacityGB, float64(totalCapacityGB))

	// 格式化容量信息
	capacityStr := fmt.Sprintf("%.2v GB", totalCapacityGB) // 总容量（GB）

	resp = &types.BucketCapacityResponse{
		Capacity:   capacityStr,
		Used:       formatBytes(bucketStat.StandardStorage),
		Percentage: percentage,
	}
	// 缓存容量信息
	marshalData, err := json.Marshal(resp)
	if err != nil {
		return nil, errors.New("marshal bucket capacity failed")
	}
	// 添加随机值（防缓存雪崩）
	// 计算缓存过期时间：距离第二天凌晨 12 点的剩余时间
	cacheExpire := timeUntilNextMidnight() + time.Duration(rand.Intn(300))*time.Second
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, marshalData, cacheExpire).Err()
	if err != nil {
		return nil, errors.New("set bucket capacity failed")
	}
	return resp, nil
}

// 提取解密操作为函数
func (l *GetBucketCapacityLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *GetBucketCapacityLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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

// 格式化字节大小为更友好的单位（KB、MB、GB 等）
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 计算使用量百分比（基于 GB）
func calculatePercentage(usedGB, totalGB float64) float64 {
	if totalGB == 0 {
		return 0
	}
	return math.Round(usedGB/totalGB*100*100) / 100
}

// 计算距离第二天凌晨 12 点的剩余时间
func timeUntilNextMidnight() time.Duration {
	now := time.Now()
	nextMidnight := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
	return nextMidnight.Sub(now)
}
