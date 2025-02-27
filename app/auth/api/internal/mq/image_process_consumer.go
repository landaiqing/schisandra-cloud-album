package mq

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/nsqio/go-nsq"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/gorm"
	"mime/multipart"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/geo_json"
	"schisandra-album-cloud-microservices/common/nsqx"
	"schisandra-album-cloud-microservices/common/storage/config"
	"strconv"
)

type NsqImageProcessConsumer struct {
	svcCtx *svc.ServiceContext
	ctx    context.Context
}

func NewImageProcessConsumer(svcCtx *svc.ServiceContext) {
	consumer := nsqx.NewNSQConsumer(constant.MQTopicImageProcess)
	consumer.AddHandler(&NsqImageProcessConsumer{
		svcCtx: svcCtx,
		ctx:    context.Background(),
	})
	err := consumer.ConnectToNSQD(svcCtx.Config.NSQ.NSQDHost)
	if err != nil {
		panic(err)
	}
}

func (c *NsqImageProcessConsumer) HandleMessage(msg *nsq.Message) error {
	if len(msg.Body) == 0 {
		return errors.New("empty message body")
	}
	var message types.FileUploadMessage
	err := json.Unmarshal(msg.Body, &message)
	if err != nil {
		return err
	}

	// 根据 GPS 信息获取地理位置信息
	country, province, city, err := c.getGeoLocation(message.Result.Latitude, message.Result.Longitude)
	if err != nil {
		return err
	}
	// 将地址信息保存到数据库
	locationId, err := c.saveFileLocationInfoToDB(message.UID, message.Result.Latitude, message.Result.Longitude, country, province, city, message.ThumbPath)
	if err != nil {
		return err
	}

	// 将文件信息存入数据库
	storageId, err := c.saveFileInfoToDB(message.UID, message.Result.Bucket, message.Result.Provider, message.FileHeader, message.Result, message.FaceID, message.FilePath, locationId, message.Result.AlbumId)
	if err != nil {
		return err
	}
	err = c.saveFileThumbnailInfoToDB(message.UID, message.ThumbPath, message.Result.ThumbW, message.Result.ThumbH, message.Result.ThumbSize, storageId)
	if err != nil {
		return err
	}
	// 删除缓存
	c.afterImageUpload(message.UID)

	return nil
}

// 根据 GPS 信息获取地理位置信息
func (c *NsqImageProcessConsumer) getGeoLocation(latitude, longitude float64) (string, string, string, error) {
	if latitude == 0.000000 || longitude == 0.000000 {
		return "", "", "", nil
	}
	country, province, city, err := geo_json.GetAddress(latitude, longitude, c.svcCtx.GeoRegionData)
	if err != nil {
		return "", "", "", errors.New("get geo location failed")
	}

	return country, province, city, nil
}

// 从缓存或数据库中获取 OSS 配置
func (c *NsqImageProcessConsumer) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*config.StorageConfig, error) {
	result, err := c.svcCtx.RedisClient.Get(c.ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.New("get oss config failed")
	}

	var ossConfig *config.StorageConfig
	if result != "" {
		var redisOssConfig model.ScaStorageConfig
		if err = json.Unmarshal([]byte(result), &redisOssConfig); err != nil {
			return nil, errors.New("unmarshal oss config failed")
		}
		return c.decryptConfig(&redisOssConfig)
	}

	// 缓存未命中，从数据库中加载
	scaOssConfig := c.svcCtx.DB.ScaStorageConfig
	dbOssConfig, err := scaOssConfig.Where(scaOssConfig.UserID.Eq(uid), scaOssConfig.Provider.Eq(provider)).First()
	if err != nil {
		return nil, err
	}

	// 缓存数据库配置
	ossConfig, err = c.decryptConfig(dbOssConfig)
	if err != nil {
		return nil, err
	}
	marshalData, err := json.Marshal(dbOssConfig)
	if err != nil {
		return nil, errors.New("marshal oss config failed")
	}
	err = c.svcCtx.RedisClient.Set(c.ctx, cacheKey, marshalData, 0).Err()
	if err != nil {
		return nil, errors.New("set oss config failed")
	}

	return ossConfig, nil
}

func (c *NsqImageProcessConsumer) classifyFile(mimeType string, isScreenshot bool) string {
	// 使用map存储MIME类型及其对应的分类
	typeMap := map[string]string{
		"image/jpeg":       "image",
		"image/png":        "image",
		"image/gif":        "gif",
		"image/bmp":        "image",
		"image/tiff":       "image",
		"image/webp":       "image",
		"video/mp4":        "video",
		"video/avi":        "video",
		"video/mpeg":       "video",
		"video/quicktime":  "video",
		"video/x-msvideo":  "video",
		"video/x-flv":      "video",
		"video/x-matroska": "video",
	}

	// 如果isScreenshot为true，则返回"screenshot"
	if isScreenshot {
		return "screenshot"
	}

	// 根据MIME类型从map中获取分类
	if classification, exists := typeMap[mimeType]; exists {
		return classification
	}

	return "unknown"
}

// 提取解密操作为函数
func (c *NsqImageProcessConsumer) decryptConfig(dbConfig *model.ScaStorageConfig) (*config.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(dbConfig.AccessKey, c.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(dbConfig.SecretKey, c.svcCtx.Config.Encrypt.Key)
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

func (c *NsqImageProcessConsumer) saveFileLocationInfoToDB(uid string, latitude float64, longitude float64, country string, province string, city string, filePath string) (int64, error) {
	if latitude == 0.000000 || longitude == 0.000000 {
		return 0, nil
	}
	locationDB := c.svcCtx.DB.ScaStorageLocation
	storageLocations, err := locationDB.Where(locationDB.UserID.Eq(uid), locationDB.Province.Eq(province), locationDB.City.Eq(city)).First()
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	if storageLocations == nil {
		locationInfo := model.ScaStorageLocation{
			UserID:     uid,
			Country:    country,
			City:       city,
			Province:   province,
			Latitude:   fmt.Sprintf("%f", latitude),
			Longitude:  fmt.Sprintf("%f", longitude),
			CoverImage: filePath,
		}
		err = locationDB.Create(&locationInfo)
		if err != nil {
			return 0, err
		}
		return 0, nil
	}
	return storageLocations.ID, nil
}

func (c *NsqImageProcessConsumer) saveFileThumbnailInfoToDB(uid string, filePath string, width, height float64, size float64, storageId int64) error {
	storageThumb := c.svcCtx.DB.ScaStorageThumb
	storageThumbInfo := &model.ScaStorageThumb{
		UserID:    uid,
		ThumbPath: filePath,
		ThumbW:    width,
		ThumbH:    height,
		ThumbSize: size,
		InfoID:    storageId,
	}
	err := storageThumb.Create(storageThumbInfo)
	if err != nil {
		logx.Error(err)
		return errors.New("create storage thumb failed")
	}
	return nil
}

// 将 EXIF 和文件信息存入数据库
func (c *NsqImageProcessConsumer) saveFileInfoToDB(uid, bucket, provider string, header *multipart.FileHeader, result types.File, faceId int64, filePath string, locationID, albumId int64) (int64, error) {
	tx := c.svcCtx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果有panic发生，回滚事务
			logx.Errorf("transaction rollback: %v", r)
		}
	}()
	typeName := c.classifyFile(result.FileType, result.IsScreenshot)
	scaStorageInfo := &model.ScaStorageInfo{
		UserID:     uid,
		Provider:   provider,
		Bucket:     bucket,
		FileName:   header.Filename,
		FileSize:   strconv.FormatInt(header.Size, 10),
		FileType:   result.FileType,
		Path:       filePath,
		FaceID:     faceId,
		Type:       typeName,
		Width:      result.Width,
		Height:     result.Height,
		LocationID: locationID,
		AlbumID:    albumId,
	}
	err := tx.ScaStorageInfo.Create(scaStorageInfo)
	if err != nil {
		tx.Rollback()
		return 0, errors.New("create storage info failed")
	}
	scaStorageExtra := &model.ScaStorageExtra{
		UserID:    uid,
		InfoID:    scaStorageInfo.ID,
		Landscape: result.Landscape,
		Tag:       result.TagName,
		IsAnime:   strconv.FormatBool(result.IsAnime),
		Category:  result.TopCategory,
	}
	err = tx.ScaStorageExtra.Create(scaStorageExtra)
	if err != nil {
		tx.Rollback()
		return 0, errors.New("create storage extra failed")
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return 0, errors.New("commit failed")
	}
	return scaStorageInfo.ID, nil
}

// 在UploadImageLogic或其他需要使缓存失效的逻辑中添加：
func (c *NsqImageProcessConsumer) afterImageUpload(uid string) {
	// 删除缓存
	keyPattern := fmt.Sprintf("%s%s:%s", constant.ImageCachePrefix, uid, "*")
	// 获取所有匹配的键
	keys, err := c.svcCtx.RedisClient.Keys(c.ctx, keyPattern).Result()
	if err != nil {
		logx.Errorf("获取缓存键 %s 失败: %v", keyPattern, err)
	}
	// 如果没有匹配的键，直接返回
	if len(keys) == 0 {
		logx.Infof("没有找到匹配的缓存键: %s", keyPattern)
		return
	}
	// 删除所有匹配的键
	if err := c.svcCtx.RedisClient.Del(c.ctx, keys...).Err(); err != nil {
		logx.Errorf("删除缓存键 %s 失败: %v", keyPattern, err)

	}
}
