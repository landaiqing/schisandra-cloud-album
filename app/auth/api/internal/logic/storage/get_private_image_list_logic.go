package storage

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gorm.io/gen"
	"io"
	"math/rand"
	"net/http"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
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

type GetPrivateImageListLogic struct {
	logx.Logger
	ctx         context.Context
	svcCtx      *svc.ServiceContext
	RestyClient *resty.Client
}

func NewGetPrivateImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPrivateImageListLogic {
	return &GetPrivateImageListLogic{
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

func (l *GetPrivateImageListLogic) GetPrivateImageList(req *types.PrivateImageListRequest) (resp *types.AllImageListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	storageInfo := l.svcCtx.DB.ScaStorageInfo
	conditions := []gen.Condition{
		storageInfo.UserID.Eq(uid),
		storageInfo.Provider.Eq(req.Provider),
		storageInfo.Bucket.Eq(req.Bucket),
		storageInfo.Type.Neq(constant.ImageTypeShared),
		storageInfo.IsDisplayed.Eq(0),
		storageInfo.IsEncrypted.Eq(constant.Encrypt),
	}
	var storageInfoList []types.FileInfoResult
	err = storageInfo.Select(
		storageInfo.ID,
		storageInfo.FileName,
		storageInfo.CreatedAt,
		storageInfo.Path).
		Where(conditions...).
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

			//  生成单条缓存键（包含文件唯一标识）
			imageCacheKey := fmt.Sprintf("%s%s:%s:%s:%s:%v",
				constant.ImageCachePrefix,
				uid,
				"list",
				req.Provider,
				req.Bucket,
				dbFileInfo.ID)
			// 尝试获取单条缓存
			if cached, err := l.svcCtx.RedisClient.Get(l.ctx, imageCacheKey).Result(); err == nil {
				var meta types.ImageMeta
				if err := json.Unmarshal([]byte(cached), &meta); err == nil {
					parse, err := time.Parse("2006-01-02 15:04:05", meta.CreatedAt)
					if err == nil {
						logx.Error("Parse Time Error:", err)
						return nil
					}
					date := parse.Format("2006年1月2日 星期") + WeekdayMap[parse.Weekday()]
					value, _ := groupedImages.LoadOrStore(date, []types.ImageMeta{})
					images := value.([]types.ImageMeta)
					images = append(images, meta)
					groupedImages.Store(date, images)
					return nil
				}
			}
			weekday := WeekdayMap[dbFileInfo.CreatedAt.Weekday()]
			date := dbFileInfo.CreatedAt.Format("2006年1月2日 星期" + weekday)
			url, err := service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.Path, time.Minute*30)
			if err != nil {
				logx.Error(err)
				return err
			}
			imageBytes, err := l.DownloadAndDecrypt(l.ctx, url, uid)
			if err != nil {
				logx.Error(err)
				return err
			}
			imageData, err := l.svcCtx.XCipher.Decrypt(imageBytes, []byte(uid))
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
				URL:       base64.StdEncoding.EncodeToString(imageData),
				Width:     dbFileInfo.ThumbW,
				Height:    dbFileInfo.ThumbH,
				CreatedAt: dbFileInfo.CreatedAt.Format("2006-01-02 15:04:05"),
			})

			// 重新存储更新后的图像列表
			groupedImages.Store(date, images)

			// 缓存单条数据（24小时基础缓存 + 随机防雪崩）
			if data, err := json.Marshal(images); err == nil {
				expire := 24*time.Hour + time.Duration(rand.Intn(3600))*time.Second
				if err := l.svcCtx.RedisClient.Set(l.ctx, imageCacheKey, data, expire).Err(); err != nil {
					logx.Error("Failed to cache image meta:", err)
				}
			}
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
	return resp, nil
}

// 提取解密操作为函数
func (l *GetPrivateImageListLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
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
func (l *GetPrivateImageListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
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

func (l *GetPrivateImageListLogic) DownloadAndDecrypt(ctx context.Context, url string, uid string) ([]byte, error) {
	resp, err := l.RestyClient.R().
		SetContext(ctx).
		SetDoNotParseResponse(true). // 保持原始响应流
		Get(url)

	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.RawBody().Close()

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	// 使用缓冲区分块读取
	buf := new(bytes.Buffer)
	if _, err := io.Copy(buf, resp.RawBody()); err != nil {
		return nil, fmt.Errorf("read response body failed: %w", err)
	}

	return buf.Bytes(), nil
}
