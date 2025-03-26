package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/redis/go-redis/v9"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/storage/config"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetImageBedUploadListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetImageBedUploadListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetImageBedUploadListLogic {
	return &GetImageBedUploadListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetImageBedUploadListLogic) GetImageBedUploadList(req *types.ImageBedUploadListRequest) (resp *types.ImageBedUploadListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageImgBed := l.svcCtx.DB.ScaStorageImgBed
	imgBeds, err := storageImgBed.Where(
		storageImgBed.UserID.Eq(uid),
		storageImgBed.Provider.Eq(req.Provider),
		storageImgBed.Bucket.Eq(req.Bucket)).Find()
	if err != nil {
		return nil, err
	}

	cacheKey := constant.UserOssConfigPrefix + uid + ":" + req.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheKey, uid, req.Provider)
	if err != nil {
		return nil, errors.New("get oss config failed")
	}
	//service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	//if err != nil {
	//	return nil, errors.New("get storage failed")
	//}
	var records []types.ImageBedUploadMeta
	for _, imgBed := range imgBeds {

		records = append(records, types.ImageBedUploadMeta{
			ID:        imgBed.ID,
			FileName:  imgBed.FileName,
			FileSize:  imgBed.FileSize,
			FileType:  imgBed.FileType,
			Path:      ossConfig.Endpoint + "/" + ossConfig.BucketName + "/" + imgBed.Path,
			Thumbnail: ossConfig.Endpoint + "/" + ossConfig.BucketName + "/" + imgBed.ThumbPath,
			CreatedAt: imgBed.CreatedAt.Format("2006-01-02 15:04:05"),
			Width:     int64(imgBed.Width),
			Height:    int64(imgBed.Height),
		})
	}
	return &types.ImageBedUploadListResponse{
		Records: records,
	}, nil
}

// 提取解密操作为函数
func (l *GetImageBedUploadListLogic) decryptConfig(dbConfig *model.ScaStorageConfig) (*config.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(dbConfig.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(dbConfig.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt secret key failed")
	}
	return &config.StorageConfig{
		Provider:   dbConfig.Provider,
		Endpoint:   dbConfig.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: dbConfig.Bucket,
		Region:     dbConfig.Region,
	}, nil
}

// 从缓存或数据库中获取 OSS 配置
func (l *GetImageBedUploadListLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*config.StorageConfig, error) {
	result, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.New("get oss config failed")
	}

	var ossConfig *config.StorageConfig
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
