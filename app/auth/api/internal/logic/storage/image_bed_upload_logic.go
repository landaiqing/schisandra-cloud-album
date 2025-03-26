package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ccpwcn/kgo"
	"github.com/redis/go-redis/v9"
	"net/http"
	"path"
	"path/filepath"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/storage/config"
	"strings"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ImageBedUploadLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewImageBedUploadLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ImageBedUploadLogic {
	return &ImageBedUploadLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ImageBedUploadLogic) ImageBedUpload(r *http.Request) (resp *types.ImageBedUploadResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	formValue := r.PostFormValue("data")
	var result types.ImageBedMeta
	if err := json.Unmarshal([]byte(formValue), &result); err != nil {
		return nil, errors.New("invalid result")
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, errors.New("file not found")
	}
	defer file.Close()
	// 解析缩略图
	thumbFile, _, err := r.FormFile("thumbnail")
	if err != nil {
		return nil, errors.New("thumbnail not found")
	}
	defer thumbFile.Close()

	// 上传文件到OSS

	cacheKey := constant.UserOssConfigPrefix + uid + ":" + result.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheKey, uid, result.Provider)
	if err != nil {
		return nil, errors.New("get oss config failed")
	}
	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return nil, errors.New("get storage failed")
	}
	filePath := path.Join(
		constant.ImageBedSpace,
		constant.ImageSpace,
		uid,
		time.Now().Format("2006/01"), // 按年/月划分目录
		fmt.Sprintf("%s_%s%s", strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)), kgo.SimpleUuid(), filepath.Ext(header.Filename)),
	)
	_, err = service.UploadFileSimple(l.ctx, ossConfig.BucketName, filePath, file, map[string]string{
		"Content-Type": header.Header.Get("Content-Type")})
	// 上传缩略图到OSS
	if err != nil {
		return nil, errors.New("upload file failed")
	}
	thumbFilePath := path.Join(
		constant.ImageBedSpace,
		constant.ThumbnailSpace,
		uid,
		time.Now().Format("2006/01"), // 按年/月划分目录
		fmt.Sprintf("%s_%s%s", strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)), kgo.SimpleUuid(), filepath.Ext(header.Filename)),
	)
	_, err = service.UploadFileSimple(l.ctx, ossConfig.BucketName, thumbFilePath, thumbFile, map[string]string{
		"Content-Type": header.Header.Get("Content-Type")})
	if err != nil {
		return nil, errors.New("upload file failed")
	}

	imgBed := model.ScaStorageImgBed{
		UserID:    uid,
		Provider:  result.Provider,
		Bucket:    ossConfig.BucketName,
		Path:      filePath,
		ThumbPath: thumbFilePath,
		FileSize:  header.Size,
		FileType:  header.Header.Get("Content-Type"),
		FileName:  header.Filename,
		Width:     float64(result.Width),
		Height:    float64(result.Height),
	}
	err = l.svcCtx.DB.ScaStorageImgBed.Create(&imgBed)
	if err != nil {
		return nil, errors.New("create image bed failed")
	}

	return &types.ImageBedUploadResponse{ID: imgBed.ID}, nil
}

// 提取解密操作为函数
func (l *ImageBedUploadLogic) decryptConfig(dbConfig *model.ScaStorageConfig) (*config.StorageConfig, error) {
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
func (l *ImageBedUploadLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*config.StorageConfig, error) {
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
