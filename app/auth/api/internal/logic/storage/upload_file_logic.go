package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ccpwcn/kgo"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
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
	"schisandra-album-cloud-microservices/common/storage/config"
	"strings"
	"sync"
	"time"
)

type UploadFileLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
	wg     sync.WaitGroup
}

func NewUploadFileLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadFileLogic {
	return &UploadFileLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
		wg:     sync.WaitGroup{},
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
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	// 解析上传的缩略图
	thumbnail, _, err := l.getUploadedThumbnail(r)
	if err != nil {
		return "", err
	}
	defer thumbnail.Close()

	// 解析图片信息识别结果
	result, err := l.parseImageInfoResult(r)
	if err != nil {
		return "", err
	}

	// 使用 `errgroup.Group` 处理并发任务
	var (
		faceId        int64
		filePath      string
		minioFilePath string
	)
	g, ctx := errgroup.WithContext(context.Background())
	// 创建信号量，限制最大并发上传数（比如最多同时 5 个任务）
	sem := semaphore.NewWeighted(5)

	// 进行人脸识别
	g.Go(func() error {
		if result.FileType == "image/png" || result.FileType == "image/jpeg" {
			face, err := l.svcCtx.AiSvcRpc.FaceRecognition(l.ctx, &pb.FaceRecognitionRequest{
				Face:   data,
				UserId: uid,
			})
			if err != nil {
				return err
			}
			if face != nil {
				faceId = face.GetFaceId()
			}
		}
		return nil
	})
	// 上传文件到 OSS
	g.Go(func() error {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		defer sem.Release(1)

		// 重新创建 `multipart.File` 兼容的 `Reader`
		fileReader := struct {
			*bytes.Reader
			io.Closer
		}{
			Reader: bytes.NewReader(data),
			Closer: io.NopCloser(nil),
		}

		fileUrl, err := l.uploadFileToOSS(uid, header, fileReader, result)
		if err != nil {
			return err
		}
		filePath = fileUrl
		return nil
	})

	// 上传缩略图到 MinIO
	g.Go(func() error {
		if err := sem.Acquire(ctx, 1); err != nil {
			return err
		}
		defer sem.Release(1)

		path, err := l.uploadFileToMinio(uid, header, thumbnail, result)
		if err != nil {
			return err
		}
		minioFilePath = path
		return nil
	})

	// 等待所有 goroutine 执行完毕
	if err = g.Wait(); err != nil {
		return "", err
	}

	fileUploadMessage := &types.FileUploadMessage{
		UID:        uid,
		Data:       result,
		FaceID:     faceId,
		FileHeader: header,
		FilePath:   filePath,
		ThumbPath:  minioFilePath,
	}
	// 转换为 JSON
	messageData, err := json.Marshal(fileUploadMessage)
	if err != nil {
		return "", err
	}
	err = l.svcCtx.NSQProducer.Publish(constant.MQTopicImageProcess, messageData)
	if err != nil {
		return "", errors.New("publish message failed")
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

// 解析上传的文件
func (l *UploadFileLogic) getUploadedThumbnail(r *http.Request) (multipart.File, *multipart.FileHeader, error) {
	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		return nil, nil, errors.New("file not found")
	}
	return file, header, nil
}

// 解析图片信息结果
func (l *UploadFileLogic) parseImageInfoResult(r *http.Request) (types.File, error) {
	formValue := r.PostFormValue("data")
	var result types.File
	if err := json.Unmarshal([]byte(formValue), &result); err != nil {
		return result, errors.New("invalid result")
	}
	return result, nil
}

// 上传文件到 OSS
func (l *UploadFileLogic) uploadFileToOSS(uid string, header *multipart.FileHeader, file multipart.File, result types.File) (string, error) {
	cacheKey := constant.UserOssConfigPrefix + uid + ":" + result.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheKey, uid, result.Provider)
	if err != nil {
		return "", errors.New("get oss config failed")
	}
	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return "", errors.New("get storage failed")
	}

	objectKey := path.Join(
		"image_space",
		uid,
		time.Now().Format("2006/01"), // 按年/月划分目录
		l.classifyFile(result.FileType, result.IsScreenshot),
		fmt.Sprintf("%s_%s%s", strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)), kgo.SimpleUuid(), filepath.Ext(header.Filename)),
	)

	_, err = service.UploadFileSimple(l.ctx, ossConfig.BucketName, objectKey, file, map[string]string{
		"Content-Type": header.Header.Get("Content-Type"),
	})
	if err != nil {
		return "", errors.New("upload file failed")
	}
	//url, err := service.PresignedURL(l.ctx, ossConfig.BucketName, objectKey, time.Hour*24*7)
	//if err != nil {
	//	return "", "", errors.New("presigned url failed")
	//}
	return objectKey, nil
}

func (l *UploadFileLogic) uploadFileToMinio(uid string, header *multipart.FileHeader, file multipart.File, result types.File) (string, error) {
	objectKey := path.Join(
		uid,
		time.Now().Format("2006/01"), // 按年/月划分目录
		l.classifyFile(result.FileType, result.IsScreenshot),
		fmt.Sprintf("%s_%s%s", strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)), kgo.SimpleUuid(), filepath.Ext(header.Filename)),
	)
	exists, err := l.svcCtx.MinioClient.BucketExists(l.ctx, constant.ThumbnailBucketName)
	if err != nil || !exists {
		err = l.svcCtx.MinioClient.MakeBucket(l.ctx, constant.ThumbnailBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			return "", err
		}
	}
	// 上传到MinIO
	_, err = l.svcCtx.MinioClient.PutObject(
		l.ctx,
		constant.ThumbnailBucketName,
		objectKey,
		file,
		int64(result.ThumbSize),
		minio.PutObjectOptions{
			ContentType: result.FileType,
		},
	)
	if err != nil {
		return "", err
	}
	//reqParams := make(url.Values)
	//presignedURL, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ThumbnailBucketName, objectKey, time.Hour*24*7, reqParams)
	//if err != nil {
	//	return "", "", err
	//}
	return objectKey, nil
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
		Provider:   dbConfig.Provider,
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

func (l *UploadFileLogic) classifyFile(mimeType string, isScreenshot bool) string {
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
