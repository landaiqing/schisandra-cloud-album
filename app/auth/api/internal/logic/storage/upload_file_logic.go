package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/zeromicro/go-zero/core/logx"
	"io"
	"mime/multipart"
	"net/http"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
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
	bucket, provider, err := l.uploadFileToOSS(uid, header, file)
	if err != nil {
		return "", err
	}

	// 将 EXIF 和文件信息存入数据库
	if err = l.saveFileInfoToDB(uid, bucket, provider, header, result, originalDateTime, gpsString, locationString, exif, faceId, className); err != nil {
		return "", err
	}

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
func (l *UploadFileLogic) uploadFileToOSS(uid string, header *multipart.FileHeader, file multipart.File) (string, string, error) {
	ossConfig := l.svcCtx.DB.ScaStorageConfig
	dbConfig, err := ossConfig.Where(ossConfig.UserID.Eq(uid)).First()
	if err != nil {
		return "", "", errors.New("oss config not found")
	}

	accessKey, err := encrypt.Decrypt(dbConfig.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", "", errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(dbConfig.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return "", "", errors.New("decrypt secret key failed")
	}

	storageConfig := &config.StorageConfig{
		Provider:   dbConfig.Type,
		Endpoint:   dbConfig.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: dbConfig.Bucket,
		Region:     dbConfig.Region,
	}

	service, err := l.svcCtx.StorageManager.GetStorage(uid, storageConfig)
	if err != nil {
		return "", "", errors.New("get storage failed")
	}

	_, err = service.UploadFileSimple(l.ctx, dbConfig.Bucket, header.Filename, file, map[string]string{})
	if err != nil {
		return "", "", errors.New("upload file failed")
	}
	return dbConfig.Bucket, dbConfig.Type, nil
}

// 将 EXIF 和文件信息存入数据库
func (l *UploadFileLogic) saveFileInfoToDB(uid, bucket, provider string, header *multipart.FileHeader, result types.File, originalDateTime, gpsString, locationString string, exif map[string]interface{}, faceId int64, className string) error {
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
