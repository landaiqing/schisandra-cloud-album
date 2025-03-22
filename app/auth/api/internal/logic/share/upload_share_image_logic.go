package share

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ccpwcn/kgo"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
	"golang.org/x/sync/errgroup"
	"image"
	"path"
	"path/filepath"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/query"
	"schisandra-album-cloud-microservices/common/constant"
	"strconv"
	"time"
)

type UploadShareImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewUploadShareImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UploadShareImageLogic {
	return &UploadShareImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *UploadShareImageLogic) UploadShareImage(req *types.ShareImageRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}

	// 启动事务，确保插入操作的原子性
	tx := l.svcCtx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback() // 如果有panic发生，回滚事务
			logx.Errorf("transaction rollback: %v", r)
		}
	}()
	albumName := req.Title
	if albumName == "" {
		albumName = "快传照片"
	}
	// 创建一个相册
	album := model.ScaStorageAlbum{
		UserID:     uid,
		AlbumName:  albumName,
		CoverImage: req.Images[0].Thumbnail,
		AlbumType:  constant.AlbumTypeShared,
	}
	err = tx.ScaStorageAlbum.Create(&album)
	if err != nil {
		return "", err
	}

	var g errgroup.Group

	// 为每张图片启动一个协程
	for _, img := range req.Images {
		img := img // 确保每个协程有独立的 img 参数副本
		g.Go(func() error {
			return l.uploadImageAndRecord(tx, uid, album, img)
		})
	}

	// 等待所有任务完成并返回第一个错误
	if err = g.Wait(); err != nil {
		tx.Rollback()
		return "", err
	}

	duration, err := strconv.Atoi(req.ExpireDate)
	if err != nil {
		return "", errors.New("invalid expire date")
	}
	var expiryTime time.Time
	if duration > 0 {
		expiryTime = l.GenerateExpiryTime(time.Now(), duration)
	}
	storageShare := model.ScaStorageShare{
		UserID:         uid,
		AlbumID:        album.ID,
		InviteCode:     kgo.SimpleUuid(),
		Status:         0,
		AccessPassword: req.AccessPassword,
		VisitLimit:     req.AccessLimit,
		ValidityPeriod: int64(duration),
		ExpireTime:     expiryTime,
		ImageCount:     int64(len(req.Images)),
	}
	err = tx.ScaStorageShare.Create(&storageShare)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	// 缓存分享码
	marshal, err := json.Marshal(storageShare)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	cacheKey := constant.ImageSharePrefix + storageShare.InviteCode
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, marshal, time.Duration(duration)*time.Hour*24).Err()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	// 提交事务
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		logx.Errorf("Transaction commit failed: %v", err)
		return "", err
	}
	return storageShare.InviteCode, nil
}

func (l *UploadShareImageLogic) uploadImageAndRecord(tx *query.QueryTx, uid string, album model.ScaStorageAlbum, img types.ShareImageMeta) error {

	// 上传原始图片到存储桶
	originImage, err := base64.StdEncoding.DecodeString(img.OriginImage)
	if err != nil {
		return fmt.Errorf("base64 decode failed: %v", err)
	}
	originObjectKey := path.Join(
		uid,
		time.Now().Format("2006/01"),
		fmt.Sprintf("%s_%s%s", img.FileName, kgo.SimpleUuid(), filepath.Ext(img.FileName)),
	)
	_, err = l.svcCtx.MinioClient.PutObject(
		l.ctx,
		constant.ShareImagesBucketName,
		originObjectKey,
		bytes.NewReader(originImage),
		int64(len(originImage)),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		logx.Errorf("Failed to upload object to MinIO: %v", err)
		return err
	}

	// 获取图片信息
	width, height, size, err := l.GetImageInfo(img.OriginImage)
	if err != nil {
		return err
	}

	// 记录原始图片信息
	imageRecord := model.ScaStorageInfo{
		UserID:      uid,
		Path:        originObjectKey,
		FileName:    img.FileName,
		FileSize:    int64(size),
		FileType:    img.FileType,
		Width:       float64(width),
		Height:      float64(height),
		Type:        constant.ImageTypeShared,
		AlbumID:     album.ID,
		IsDisplayed: 1,
	}
	err = tx.ScaStorageInfo.Create(&imageRecord)
	if err != nil {
		return err
	}

	// 上传缩略图到 Minio
	thumbnail, err := base64.StdEncoding.DecodeString(img.Thumbnail)
	if err != nil {
		return fmt.Errorf("base64 decode failed: %v", err)
	}
	thumbObjectKey := path.Join(
		uid,
		time.Now().Format("2006/01"),
		l.classifyFile(img.FileType),
		fmt.Sprintf("%s_%s.jpg", time.Now().Format("20060102150405"), kgo.SimpleUuid()),
	)

	_, err = l.svcCtx.MinioClient.PutObject(
		l.ctx,
		constant.ThumbnailBucketName,
		thumbObjectKey,
		bytes.NewReader(thumbnail),
		int64(len(thumbnail)),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		logx.Errorf("Failed to upload MinIO object: %v", err)
		return err
	}

	// 记录缩略图
	thumbRecord := model.ScaStorageThumb{
		InfoID:    imageRecord.ID,
		UserID:    uid,
		ThumbPath: thumbObjectKey,
		ThumbW:    img.ThumbW,
		ThumbH:    img.ThumbH,
		ThumbSize: float64(len(thumbnail)),
	}
	err = tx.ScaStorageThumb.Create(&thumbRecord)
	if err != nil {
		return err
	}

	return nil
}

func (l *UploadShareImageLogic) GetImageInfo(base64Str string) (width, height int, size int, err error) {
	// 解码 Base64
	data, err := base64.StdEncoding.DecodeString(base64Str)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("base64 decode failed: %v", err)
	}

	// 获取图片大小
	size = len(data)

	// 解析图片宽高
	reader := bytes.NewReader(data)
	imgCfg, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("decode image config failed: %v", err)
	}

	return imgCfg.Width, imgCfg.Height, size, nil
}

// GenerateExpiryTime 函数接受当前时间和有效期（天为单位），返回过期时间
func (l *UploadShareImageLogic) GenerateExpiryTime(currentTime time.Time, durationInDays int) time.Time {
	// 创建一个持续时间对象
	duration := time.Duration(durationInDays) * 24 * time.Hour
	// 将当前时间加上持续时间，得到过期时间
	expiryTime := currentTime.Add(duration)
	return expiryTime
}

func (l *UploadShareImageLogic) classifyFile(mimeType string) string {
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
	return "other"
}
