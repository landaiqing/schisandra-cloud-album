package share

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gorm.io/gorm"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryShareImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryShareImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryShareImageLogic {
	return &QueryShareImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryShareImageLogic) QueryShareImage(req *types.QueryShareImageRequest) (resp *types.QueryShareImageResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	// 获取分享记录
	cacheKey := constant.ImageSharePrefix + req.ShareCode
	shareData, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, errors.New("share code not found")
		}
		return nil, err
	}
	var storageShare model.ScaStorageShare
	if err := json.Unmarshal([]byte(shareData), &storageShare); err != nil {
		return nil, errors.New("unmarshal share data failed")
	}

	// 验证密码
	if storageShare.AccessPassword != "" && storageShare.AccessPassword != req.AccessPassword {
		return nil, errors.New("incorrect password")
	}

	// 检查分享是否过期
	if storageShare.ExpireTime.Before(time.Now()) {
		return nil, errors.New("share link has expired")
	}

	// 检查访问限制
	if storageShare.VisitLimit > 0 {
		err = l.incrementVisitCount(req.ShareCode, storageShare.VisitLimit)
		if err != nil {
			return nil, err
		}
	}
	// 记录用户访问
	err = l.recordUserVisit(storageShare.ID, uid)
	if err != nil {
		logx.Error("Failed to record user visit:", err)
		return nil, err
	}

	// 生成缓存键（在验证通过后）
	resultCacheKey := constant.ImageListPrefix + req.ShareCode + ":" + req.AccessPassword

	// 尝试从缓存中获取结果
	cachedResult, err := l.svcCtx.RedisClient.Get(l.ctx, resultCacheKey).Result()
	if err == nil {
		// 缓存命中，直接返回缓存结果
		var cachedResponse types.QueryShareImageResponse
		if err := json.Unmarshal([]byte(cachedResult), &cachedResponse); err == nil {
			return &cachedResponse, nil
		}
		logx.Error("Failed to unmarshal cached result:", err)
	} else if !errors.Is(err, redis.Nil) {
		// 如果 Redis 查询出错（非缓存未命中），记录错误并继续回源查询
		logx.Error("Failed to get cached result from Redis:", err)
	}
	// 缓存未命中，执行回源查询逻辑
	resp, err = l.queryShareImageFromSource(&storageShare)
	if err != nil {
		return nil, err
	}

	// 将查询结果缓存到 Redis
	respBytes, err := json.Marshal(resp)
	if err != nil {
		logx.Error("Failed to marshal response for caching:", err)
	} else {
		// 设置缓存，过期时间为 5 分钟
		err = l.svcCtx.RedisClient.Set(l.ctx, resultCacheKey, respBytes, 5*time.Minute).Err()
		if err != nil {
			logx.Error("Failed to cache result in Redis:", err)
		}
	}

	return resp, nil
}

func (l *QueryShareImageLogic) queryShareImageFromSource(storageShare *model.ScaStorageShare) (resp *types.QueryShareImageResponse, err error) {
	// 查询相册图片列表
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageThumb := l.svcCtx.DB.ScaStorageThumb
	var storageInfoList []types.ShareFileInfoResult
	err = storageInfo.Select(
		storageInfo.ID,
		storageInfo.FileName,
		storageInfo.CreatedAt,
		storageInfo.Provider,
		storageInfo.Bucket,
		storageInfo.Path,
		storageThumb.ThumbPath,
		storageThumb.ThumbW,
		storageThumb.ThumbH,
		storageThumb.ThumbSize).
		LeftJoin(storageThumb, storageInfo.ThumbID.EqCol(storageThumb.ID)).
		Where(
			storageInfo.Type.Eq(constant.ImageTypeShared),
			storageInfo.AlbumID.Eq(storageShare.AlbumID)).
		Order(storageInfo.CreatedAt.Desc()).Scan(&storageInfoList)
	if err != nil {
		return nil, err
	}

	// 使用 errgroup 和 semaphore 并发处理图片信息
	var ResultList []types.ShareImageListMeta
	g, ctx := errgroup.WithContext(l.ctx)
	sem := semaphore.NewWeighted(10) // 限制并发数为 10

	for _, imgInfo := range storageInfoList {
		imgInfo := imgInfo // 创建局部变量，避免闭包问题
		if err := sem.Acquire(ctx, 1); err != nil {
			return nil, err
		}
		g.Go(func() error {
			defer sem.Release(1)
			// 加载用户oss配置信息
			cacheOssConfigKey := constant.UserOssConfigPrefix + storageShare.UserID + ":" + imgInfo.Provider
			ossConfig, err := l.getOssConfigFromCacheOrDb(cacheOssConfigKey, storageShare.UserID, imgInfo.Provider)
			if err != nil {
				return err
			}

			service, err := l.svcCtx.StorageManager.GetStorage(storageShare.UserID, ossConfig)
			if err != nil {
				return errors.New("get storage failed")
			}
			ossURL, err := service.PresignedURL(ctx, ossConfig.BucketName, imgInfo.Path, 30*time.Minute)
			if err != nil {
				return errors.New("get presigned url failed")
			}
			reqParams := make(url.Values)
			presignedURL, err := l.svcCtx.MinioClient.PresignedGetObject(ctx, constant.ThumbnailBucketName, imgInfo.ThumbPath, 30*time.Minute, reqParams)
			if err != nil {
				return errors.New("get presigned thumbnail url failed")
			}
			ResultList = append(ResultList, types.ShareImageListMeta{
				ID:        imgInfo.ID,
				FileName:  imgInfo.FileName,
				ThumbH:    imgInfo.ThumbH,
				ThumbW:    imgInfo.ThumbW,
				ThumbSize: imgInfo.ThumbSize,
				CreatedAt: imgInfo.CreatedAt.Format(constant.TimeFormat),
				URL:       ossURL,
				Thumbnail: presignedURL.String(),
			})
			return nil
		})
	}

	// 等待所有并发任务完成
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return &types.QueryShareImageResponse{
		List: ResultList}, nil
}

func (l *QueryShareImageLogic) recordUserVisit(shareID int64, userID string) error {
	// 查询是否已经存在该用户对该分享的访问记录
	var visitRecord model.ScaStorageShareVisit
	scaStorageShareVisit := l.svcCtx.DB.ScaStorageShareVisit
	_, err := scaStorageShareVisit.
		Where(scaStorageShareVisit.ShareID.Eq(shareID), scaStorageShareVisit.UserID.Eq(userID)).
		First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// 如果记录不存在，创建新的访问记录
			visitRecord = model.ScaStorageShareVisit{
				UserID:  userID,
				ShareID: shareID,
				Views:   1,
			}
			err = l.svcCtx.DB.ScaStorageShareVisit.Create(&visitRecord)
			if err != nil {
				return errors.New("failed to create visit record")
			}
			return nil
		}
		return errors.New("failed to query visit record")
	}

	// 如果记录存在，增加访问次数
	info, err := scaStorageShareVisit.
		Where(scaStorageShareVisit.UserID.Eq(userID), scaStorageShareVisit.ShareID.Eq(shareID)).
		Update(scaStorageShareVisit.Views, scaStorageShareVisit.Views.Add(1))
	if err != nil {
		return errors.New("failed to update visit record")
	}
	if info.RowsAffected == 0 {
		return errors.New("failed to update visit record")
	}

	return nil
}
func (l *QueryShareImageLogic) incrementVisitCount(shareCode string, limit int64) error {
	// Redis 键值
	cacheKey := constant.ImageShareVisitPrefix + shareCode
	currentVisitCount, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Int64()
	if err != nil && !errors.Is(err, redis.Nil) {
		return err
	}

	// 如果访问次数超过限制，返回错误
	if currentVisitCount >= limit {
		return errors.New("access limit reached")
	}

	// 增加访问次数
	err = l.svcCtx.RedisClient.Incr(l.ctx, cacheKey).Err()
	if err != nil {
		return err
	}

	return nil
}

// 提取解密操作为函数
func (l *QueryShareImageLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *QueryShareImageLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
