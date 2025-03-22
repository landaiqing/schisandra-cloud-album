package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
	"gorm.io/gen"
	"gorm.io/gorm"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/captcha/verify"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"schisandra-album-cloud-microservices/common/utils"
	"sort"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetPrivateImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetPrivateImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetPrivateImageListLogic {
	return &GetPrivateImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetPrivateImageListLogic) GetPrivateImageList(req *types.PrivateImageListRequest) (resp *types.AllImageListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	captcha := verify.VerifyBasicTextCaptcha(req.Dots, req.Key, l.svcCtx.RedisClient, l.ctx)
	if !captcha {
		return nil, errors.New("验证错误")
	}
	if req.Password == "" {
		return nil, errors.New("密码不能为空")
	}
	authUser := l.svcCtx.DB.ScaAuthUser
	userInfo, err := authUser.
		Select(authUser.UID, authUser.Password).
		Where(authUser.UID.Eq(uid)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if userInfo == nil {
		return nil, errors.New("密码错误")
	}
	if !utils.Verify(userInfo.Password, req.Password) {
		return nil, errors.New("密码错误")
	}

	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageThumb := l.svcCtx.DB.ScaStorageThumb
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
		storageThumb.ThumbPath,
		storageInfo.Path,
		storageThumb.ThumbW,
		storageThumb.ThumbH,
		storageThumb.ThumbSize,
	).
		LeftJoin(storageThumb, storageInfo.ID.EqCol(storageThumb.InfoID)).
		Where(conditions...).
		Order(storageInfo.CreatedAt.Desc()).Scan(&storageInfoList)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
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

			thumbnailUrl, err := service.PresignedURL(l.ctx, ossConfig.BucketName, dbFileInfo.ThumbPath, time.Minute*30)
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
				Thumbnail: thumbnailUrl,
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
