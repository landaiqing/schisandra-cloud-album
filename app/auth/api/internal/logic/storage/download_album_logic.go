package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DownloadAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDownloadAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DownloadAlbumLogic {
	return &DownloadAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DownloadAlbumLogic) DownloadAlbum(req *types.DownloadAlbumRequest) (resp *types.DownloadAlbumResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageInfos, err := storageInfo.Where(storageInfo.UserID.Eq(uid), storageInfo.AlbumID.Eq(req.ID),
		storageInfo.Provider.Eq(req.Provider), storageInfo.Bucket.Eq(req.Bucket)).Find()
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
	// 并发生成预签名URL
	urls := make([]string, len(storageInfos))
	g, ctx := errgroup.WithContext(l.ctx)
	for i := range storageInfos {
		i := i
		file := storageInfos[i]
		g.Go(func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				url, err := service.PresignedURL(ctx, ossConfig.BucketName, file.Path, 5*time.Minute)
				if err != nil {
					return fmt.Errorf("failed to generate URL for %s: %w", file.Path, err)
				}
				urls[i] = url
				return nil
			}
		})
	}

	if err = g.Wait(); err != nil {
		return nil, fmt.Errorf("failed to generate URLs: %w", err)
	}

	return &types.DownloadAlbumResponse{
		Records: urls,
	}, nil
}

// 提取解密操作为函数
func (l *DownloadAlbumLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *DownloadAlbumLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
