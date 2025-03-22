package storage

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/hybrid_encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"schisandra-album-cloud-microservices/common/utils"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPrivateImageUrlLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	RestyClient *resty.Client
}

func NewGetPrivateImageUrlLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPrivateImageUrlLogic {
	return &GetPrivateImageUrlLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		RestyClient: resty.New().
			SetTimeout(30 * time.Second).          // 总超时时间
			SetRetryCount(3).                      // 重试次数
			SetRetryWaitTime(5 * time.Second).     // 重试等待时间
			SetRetryMaxWaitTime(30 * time.Second). // 最大重试等待
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return r.StatusCode() == http.StatusTooManyRequests ||
					err != nil ||
					r.StatusCode() >= 500
			}),
	}
}

// 修改函数签名和实现
func (l *GetPrivateImageUrlLogic) GetPrivateImageUrl(req *types.SinglePrivateImageRequest) (string, error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}

	// 构建缓存key
	cacheKey := fmt.Sprintf("%s%s:%s:%v", constant.ImageCachePrefix, uid, "encrypted", req.ID)

	// 检查缓存
	if cachedData, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result(); err == nil {
		return cachedData, nil
	}

	storageInfo := l.svcCtx.DB.ScaStorageInfo
	authUser := l.svcCtx.DB.ScaAuthUser
	var result struct {
		ID       int64  `json:"id"`
		Path     string `json:"path"`
		Password string `json:"password"`
		FileType string `json:"file_type"`
	}
	err := storageInfo.
		Select(
			storageInfo.ID,
			storageInfo.Path,
			storageInfo.FileType,
			authUser.Password).
		LeftJoin(authUser, authUser.UID.EqCol(storageInfo.UserID)).
		Where(storageInfo.ID.Eq(req.ID), storageInfo.UserID.Eq(uid),
			storageInfo.IsEncrypted.Eq(constant.Encrypt), storageInfo.Provider.Eq(req.Provider),
			storageInfo.Bucket.Eq(req.Bucket), authUser.UID.Eq(uid)).
		Group(storageInfo.ID).Scan(&result)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}
	if result.ID == 0 {
		return "", errors.New("image not found")
	}
	verify := utils.Verify(result.Password, req.Password)
	if !verify {
		return "", errors.New("invalid password")
	}

	// 加载用户oss配置信息
	cacheOssConfigKey := constant.UserOssConfigPrefix + uid + ":" + req.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheOssConfigKey, uid, req.Provider)
	if err != nil {
		return "", err
	}

	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return "", errors.New("get storage failed")
	}
	url, err := service.PresignedURL(l.ctx, ossConfig.BucketName, result.Path, time.Minute*15)
	if err != nil {
		logx.Error(err)
		return "", errors.New("get private image url failed")
	}

	resp, err := l.RestyClient.R().
		SetContext(l.ctx).
		SetDoNotParseResponse(true). // 保持原始响应流
		Get(url)
	if err != nil {
		return "", fmt.Errorf("download private image failed: %w", err)
	}
	defer resp.RawBody().Close()

	body, err := io.ReadAll(resp.RawBody())
	if err != nil {
		return "", err
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	privateKeyPath := filepath.Join(dir, l.svcCtx.Config.Encrypt.PrivateKey)
	privateKey, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", err
	}

	pem, err := hybrid_encrypt.ImportPrivateKeyPEM(privateKey)
	if err != nil {
		return "", err
	}
	image, err := hybrid_encrypt.DecryptImage(pem, body)
	if err != nil {
		return "", err
	}
	base64Str := base64.StdEncoding.EncodeToString(image)

	// 设置缓存，过期时间设为12小时
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, base64Str, 12*time.Hour).Err()
	if err != nil {
		logx.Errorf("cache private image failed: %v", err)
	}

	return base64Str, nil
}

// 提取解密操作为函数
func (l *GetPrivateImageUrlLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *GetPrivateImageUrlLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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
