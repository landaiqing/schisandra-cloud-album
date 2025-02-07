package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ccpwcn/kgo"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"path/filepath"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	"schisandra-album-cloud-microservices/common/gao_map"
	"schisandra-album-cloud-microservices/common/storage/config"
	"strconv"
	"strings"
	"time"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadFileLogic) UploadFile(r *http.Request) (resp string, err error) {
	// 获取用户 ID
	uid, err := l.getUserID()
	if err != nil {
		return "", err
	}

	// 解析上传的文件
	file, header, err := l.getUploadedFile(r)
	if err != nil {
		return "", err
	}

	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// 解析 AI 识别结果
	result, err := l.parseAIRecognitionResult(r)
	if err != nil {
		return "", err
	}
	var faceId int64 = 0
	var className string
	bytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	// 人脸识别
	if result.FileType == "image/png" || result.FileType == "image/jpeg" {
		face, err := l.svcCtx.AiSvcRpc.FaceRecognition(l.ctx, &pb.FaceRecognitionRequest{Face: bytes, UserId: uid})
		if err != nil {
			return "", err
		}
		if face != nil {
			faceId = face.GetFaceId()
		}

		// 图像分类
		classification, err := l.svcCtx.AiSvcRpc.TfClassification(l.ctx, &pb.TfClassificationRequest{Image: bytes})
		if err != nil {
			return "", err
		}
		className = classification.GetClassName()
	}

	// 解析 EXIF 信息
	exif, err := l.parseExifData(result.Exif)
	if err != nil {
		return "", err
	}

	// 提取拍摄时间
	originalDateTime, err := l.extractOriginalDateTime(exif)
	if err != nil {
		return "", err
	}

	// 提取 GPS 信息
	latitude, longitude := l.extractGPSCoordinates(exif)

	// 根据 GPS 信息获取地理位置信息
	locationString, gpsString, err := l.getGeoLocation(latitude, longitude)
	if err != nil {
		return "", err
	}

	// 上传文件到 OSS
	// 重新设置文件指针到文件开头
	if _, err = file.Seek(0, 0); err != nil {
		return "", err
	}
	bucket, provider, filePath, err := l.uploadFileToOSS(uid, header, file, result)
	if err != nil {
		return "", err
	}

	// 将 EXIF 和文件信息存入数据库
	if err = l.saveFileInfoToDB(uid, bucket, provider, header, result, originalDateTime, gpsString, locationString, exif, faceId, className, filePath); err != nil {
		return "", err
	}

	// 删除缓存
	l.afterImageUpload(uid, provider, bucket)

	return "success", nil
}

// 获取用户 ID
func (l *UploadFileLogic) getUserID() (string, error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	return uid, nil
}

// 在UploadImageLogic或其他需要使缓存失效的逻辑中添加：
func (l *UploadFileLogic) afterImageUpload(uid, provider, bucket string) {
	// 构造所有可能的缓存键组合（sort为true/false）
	keysToDelete := []string{
		fmt.Sprintf("%s%s:%s:%s:true", constant.ImageListPrefix, uid, provider, bucket),
		fmt.Sprintf("%s%s:%s:%s:false", constant.ImageListPrefix, uid, provider, bucket),
	}

	// 批量删除缓存
	for _, key := range keysToDelete {
		if err := l.svcCtx.RedisClient.Del(l.ctx, key).Err(); err != nil {
			logx.Errorf("Failed to delete cache key %s: %v", key, err)
		}
	}
}

// 解析上传的文件
func (l *UploadFileLogic) getUploadedFile(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := r.FormFile("file")
	if err != nil {
		return nil, nil, errors.New("file not found")
	}
	return file, header, nil
}

// 解析 AI 识别结果
func (l *UploadFileLogic) parseAIRecognitionResult(r *http.Request) (types.File, error) {
	formValue := r.PostFormValue("data")
	var result types.File
	if err := json.Unmarshal([]byte(formValue), &result); err != nil {
		return result, errors.New("invalid result")
	}
	return result, nil
}

// 解析 EXIF 数据
func (l *UploadFileLogic) parseExifData(exifData interface{}) (map[string]interface{}, error) {
	marshaledExif, err := json.Marshal(exifData)
	if err != nil {
		return nil, errors.New("invalid exif")
	}

	var exif map[string]interface{}
	if err = json.Unmarshal(marshaledExif, &exif); err != nil {
		return nil, errors.New("invalid exif")
	}
	return exif, nil
}

// 提取拍摄时间
func (l *UploadFileLogic) extractOriginalDateTime(exif map[string]interface{}) (string, error) {
	if dateTimeOriginal, ok := exif["DateTimeOriginal"].(string); ok {
		parsedTime, err := time.Parse(time.RFC3339, dateTimeOriginal)
		if err == nil {
			return parsedTime.Format("2006-01-02 15:04:05"), nil
		}
	}
	return "", nil
}

// 提取 GPS 信息
func (l *UploadFileLogic) extractGPSCoordinates(exif map[string]interface{}) (float64, float64) {
	var latitude, longitude float64
	if lat, ok := exif["latitude"].(float64); ok {
		latitude = lat
	}
	if long, ok := exif["longitude"].(float64); ok {
		longitude = long
	}
	return latitude, longitude
}

// 根据 GPS 信息获取地理位置信息
func (l *UploadFileLogic) getGeoLocation(latitude, longitude float64) (string, string, error) {
	if latitude == 0 || longitude == 0 {
		return "", "", nil
	}

	gpsString := fmt.Sprintf("[%f,%f]", latitude, longitude)
	request := gao_map.ReGeoRequest{Location: fmt.Sprintf("%f,%f", latitude, longitude)}

	location, err := l.svcCtx.GaoMap.Location.ReGeo(&request)
	if err != nil {
		return "", "", errors.New("regeo failed")
	}

	addressInfo := map[string]string{}
	if location.ReGeoCode.AddressComponent.Country != "" {
		addressInfo["county"] = location.ReGeoCode.AddressComponent.Country
	}
	if location.ReGeoCode.AddressComponent.Province != "" {
		addressInfo["province"] = location.ReGeoCode.AddressComponent.Province
	}
	if location.ReGeoCode.AddressComponent.City != "" {
		addressInfo["city"] = location.ReGeoCode.AddressComponent.City.(string)
	}
	if location.ReGeoCode.AddressComponent.District != "" {
		addressInfo["district"] = location.ReGeoCode.AddressComponent.District.(string)
	}
	if location.ReGeoCode.AddressComponent.Township != "" {
		addressInfo["township"] = location.ReGeoCode.AddressComponent.Township
	}

	locationString := ""
	if len(addressInfo) > 0 {
		addressJSON, err := json.Marshal(addressInfo)
		if err != nil {
			return "", "", errors.New("marshal address info failed")
		}
		locationString = string(addressJSON)
	}

	return locationString, gpsString, nil
}

// 上传文件到 OSS
func (l *UploadFileLogic) uploadFileToOSS(uid string, header *multipart.FileHeader, file multipart.File, result types.File) (string, string, string, error) {
	cacheKey := constant.UserOssConfigPrefix + uid + ":" + result.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheKey, uid, result.Provider)
	if err != nil {
		return "", "", "", errors.New("get oss config failed")
	}
	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return "", "", "", errors.New("get storage failed")
	}

	objectKey := path.Join(
		uid,
		time.Now().Format("2006/01"), // 按年/月划分目录
		fmt.Sprintf("%s_%s%s", strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)), kgo.SimpleUuid(), filepath.Ext(header.Filename)),
	)

	_, err = service.UploadFileSimple(l.ctx, ossConfig.BucketName, objectKey, file, map[string]string{
		"Content-Type": header.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", "", "", errors.New("upload file failed")
	}
	return ossConfig.BucketName, ossConfig.Provider, objectKey, nil
}

// 将 EXIF 和文件信息存入数据库
func (l *UploadFileLogic) saveFileInfoToDB(uid, bucket, provider string, header *multipart.FileHeader, result types.File, originalDateTime, gpsString, locationString string, exif map[string]interface{}, faceId int64, className, filePath string) error {
	exifJSON, err := json.Marshal(exif)
	if err != nil {
		return errors.New("marshal exif failed")
	}
	var landscape string
	if result.Landscape != "none" {
		landscape = result.Landscape
	}
	scaStorageInfo := &model.ScaStorageInfo{
		UserID:       uid,
		Provider:     provider,
		Bucket:       bucket,
		FileName:     header.Filename,
		FileSize:     strconv.FormatInt(header.Size, 10),
		FileType:     result.FileType,
		Path:         filePath,
		Landscape:    landscape,
		Objects:      strings.Join(result.ObjectArray, ", "),
		Anime:        strconv.FormatBool(result.IsAnime),
		Category:     result.TopCategory,
		Screenshot:   strconv.FormatBool(result.IsScreenshot),
		OriginalTime: originalDateTime,
		Gps:          gpsString,
		Location:     locationString,
		Exif:         string(exifJSON),
		FaceID:       faceId,
		Tags:         className,
	}

	err = l.svcCtx.DB.ScaStorageInfo.Create(scaStorageInfo)
	if err != nil {
		return errors.New("create storage info failed")
	}
	return nil
}

// 提取解密操作为函数
func (l *UploadFileLogic) decryptConfig(dbConfig *model.ScaStorageConfig) (*config.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(dbConfig.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(dbConfig.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt secret key failed")
	}
	return &config.StorageConfig{
		Provider:   dbConfig.Type,
		Endpoint:   dbConfig.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: dbConfig.Bucket,
		Region:     dbConfig.Region,
	}, nil
}

// 从缓存或数据库中获取 OSS 配置
func (l *UploadFileLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*config.StorageConfig, error) {
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
	dbOssConfig, err := scaOssConfig.Where(scaOssConfig.UserID.Eq(uid), scaOssConfig.Type.Eq(provider)).First()
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
