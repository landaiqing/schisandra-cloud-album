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
	"time"
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
	country, province, city, err := c.getGeoLocation(message.Data.Latitude, message.Data.Longitude)
	if err != nil {
		return err
	}
	// 将地址信息保存到数据库
	locationId, err := c.saveFileLocationInfoToDB(message.UID, message.Data.Provider, message.Data.Bucket, message.Data.Latitude, message.Data.Longitude, country, province, city, message.ThumbPath)
	if err != nil {
		return err
	}

	thumbnailId, err := c.saveFileThumbnailInfoToDB(message.UID, message.ThumbPath, message.Data.ThumbW, message.Data.ThumbH, message.Data.ThumbSize)

	// 将文件信息存入数据库
	id, err := c.saveFileInfoToDB(message.UID, message.Data.Bucket, message.Data.Provider, message.FileHeader, message.Data, locationId, message.FaceID, message.FilePath, thumbnailId)
	if err != nil {
		return err
	}
	// 删除缓存
	c.afterImageUpload(message.UID, message.Data.Provider, message.Data.Bucket)

	// redis 保存最近7天上传的文件列表
	err = c.saveRecentFileList(message.UID, message.PresignedURL, id, message.Data, message.FileHeader.Filename)
	if err != nil {
		return err
	}

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

	// 根据MIME类型从map中获取分类
	if classification, exists := typeMap[mimeType]; exists {
		return classification
	}

	// 如果isScreenshot为true，则返回"screenshot"
	if isScreenshot {
		return "screenshot"
	}
	return "unknown"
}

// 保存最近7天上传的文件列表
func (c *NsqImageProcessConsumer) saveRecentFileList(uid, url string, id int64, result types.File, filename string) error {

	redisKey := constant.ImageRecentPrefix + uid + ":" + strconv.FormatInt(id, 10)
	imageMeta := types.ImageMeta{
		ID:        id,
		URL:       url,
		FileName:  filename,
		Width:     result.Width,
		Height:    result.Height,
		CreatedAt: time.Now().Format("2006-01-02 15:04:05"),
	}
	marshal, err := json.Marshal(imageMeta)
	if err != nil {
		logx.Error(err)
		return errors.New("marshal image meta failed")
	}
	err = c.svcCtx.RedisClient.Set(c.ctx, redisKey, marshal, time.Hour*24*7).Err()
	if err != nil {
		logx.Error(err)
		return errors.New("save recent file list failed")
	}
	return nil
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

func (c *NsqImageProcessConsumer) saveFileLocationInfoToDB(uid string, provider string, bucket string, latitude float64, longitude float64, country string, province string, city string, filePath string) (int64, error) {
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
			Provider:   provider,
			Bucket:     bucket,
			UserID:     uid,
			Country:    country,
			City:       city,
			Province:   province,
			Latitude:   fmt.Sprintf("%f", latitude),
			Longitude:  fmt.Sprintf("%f", longitude),
			Total:      1,
			CoverImage: filePath,
		}
		err = locationDB.Create(&locationInfo)
		if err != nil {
			return 0, err
		}
		return locationInfo.ID, nil
	} else {
		info, err := locationDB.Where(locationDB.ID.Eq(storageLocations.ID), locationDB.UserID.Eq(uid)).UpdateColumnSimple(locationDB.Total.Add(1), locationDB.CoverImage.Value(filePath))
		if err != nil {
			return 0, err
		}
		if info.RowsAffected == 0 {
			return 0, errors.New("update location failed")
		}
		return storageLocations.ID, nil
	}
}

func (c *NsqImageProcessConsumer) saveFileThumbnailInfoToDB(uid string, filePath string, width, height float64, size float64) (int64, error) {
	storageThumb := c.svcCtx.DB.ScaStorageThumb
	storageThumbInfo := &model.ScaStorageThumb{
		UserID:    uid,
		ThumbPath: filePath,
		ThumbW:    width,
		ThumbH:    height,
		ThumbSize: size,
	}
	err := storageThumb.Create(storageThumbInfo)
	if err != nil {
		logx.Error(err)
		return 0, errors.New("create storage thumb failed")
	}
	return storageThumbInfo.ID, nil
}

// 将 EXIF 和文件信息存入数据库
func (c *NsqImageProcessConsumer) saveFileInfoToDB(uid, bucket, provider string, header *multipart.FileHeader, result types.File, locationId, faceId int64, filePath string, thumbnailId int64) (int64, error) {

	typeName := c.classifyFile(result.FileType, result.IsScreenshot)
	scaStorageInfo := &model.ScaStorageInfo{
		UserID:     uid,
		Provider:   provider,
		Bucket:     bucket,
		FileName:   header.Filename,
		FileSize:   strconv.FormatInt(header.Size, 10),
		FileType:   result.FileType,
		Path:       filePath,
		Landscape:  result.Landscape,
		Tag:        result.TagName,
		IsAnime:    strconv.FormatBool(result.IsAnime),
		Category:   result.TopCategory,
		LocationID: locationId,
		FaceID:     faceId,
		Type:       typeName,
		Width:      result.Width,
		Height:     result.Height,
		ThumbID:    thumbnailId,
	}

	err := c.svcCtx.DB.ScaStorageInfo.Create(scaStorageInfo)
	if err != nil {
		return 0, errors.New("create storage info failed")
	}
	return scaStorageInfo.ID, nil
}

// 在UploadImageLogic或其他需要使缓存失效的逻辑中添加：
func (c *NsqImageProcessConsumer) afterImageUpload(uid, provider, bucket string) {
	for _, sort := range []bool{true, false} {
		key := fmt.Sprintf("%s%s:%s:%s:%v", constant.ImageListPrefix, uid, provider, bucket, sort)
		if err := c.svcCtx.RedisClient.Del(c.ctx, key).Err(); err != nil {
			logx.Errorf("删除缓存键 %s 失败: %v", key, err)
		}
	}
}
