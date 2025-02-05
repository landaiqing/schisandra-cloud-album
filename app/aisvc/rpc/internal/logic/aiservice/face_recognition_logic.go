package aiservicelogic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/Kagami/go-face"
	"github.com/ccpwcn/kgo"
	"github.com/minio/minio-go/v7"
	"github.com/zeromicro/go-zero/core/logx"
	"image"
	"image/jpeg"
	_ "image/png"
	"path"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/model"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"
	"schisandra-album-cloud-microservices/common/constant"
	"strconv"
	"sync"
	"time"
)

type FaceRecognitionLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	directoryCache sync.Map
	wg             sync.WaitGroup
	mu             sync.Mutex
}

func NewFaceRecognitionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *FaceRecognitionLogic {
	return &FaceRecognitionLogic{
		ctx:            ctx,
		svcCtx:         svcCtx,
		Logger:         logx.WithContext(ctx),
		directoryCache: sync.Map{},
		wg:             sync.WaitGroup{},
		mu:             sync.Mutex{},
	}
}

// FaceRecognition 人脸识别
func (l *FaceRecognitionLogic) FaceRecognition(in *pb.FaceRecognitionRequest) (*pb.FaceRecognitionResponse, error) {
	toJPEG, err := l.ConvertImageToJPEG(in.GetFace())
	if err != nil {
		return nil, err
	}
	if toJPEG == nil {
		return nil, nil
	}
	// 提取人脸特征
	faceFeatures, err := l.svcCtx.FaceRecognizer.RecognizeSingle(toJPEG)
	if err != nil {
		return nil, err
	}
	if faceFeatures == nil {
		return nil, nil
	}

	hashKey := constant.FaceVectorPrefix + in.GetUserId()
	// 从 Redis 加载人脸数据
	samples, ids, err := l.loadFacesFromRedisHash(hashKey)
	if err != nil {
		return nil, fmt.Errorf("failed to query Redis: %v", err)
	}
	// 如果缓存中没有数据，则查询数据库
	if len(samples) == 0 {
		samples, ids, err = l.loadExistingFaces(in.GetUserId())
		if err != nil {
			return nil, err
		}

		// 如果数据库也没有数据，直接保存当前人脸
		if len(samples) == 0 || len(ids) == 0 {
			return l.saveNewFace(in, faceFeatures, hashKey)
		}

		// 将数据写入 Redis
		err = l.cacheFacesToRedisHash(hashKey, samples, ids)
		if err != nil {
			return nil, fmt.Errorf("failed to cache faces to Redis: %v", err)
		}
	}

	// 设置人脸特征
	l.svcCtx.FaceRecognizer.SetSamples(samples, ids)

	// 人脸分类
	classify := l.svcCtx.FaceRecognizer.ClassifyThreshold(faceFeatures.Descriptor, 0.3)
	if classify > 0 {
		return &pb.FaceRecognitionResponse{
			FaceId: int64(classify),
		}, nil
	}

	// 如果未找到匹配的人脸，则保存为新样本
	return l.saveNewFace(in, faceFeatures, hashKey)
}

func (l *FaceRecognitionLogic) ConvertImageToJPEG(imageData []byte) ([]byte, error) {

	// 解码图片
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	// 如果已经是 JPEG 格式，则直接返回原数据
	if format == "jpeg" {
		return imageData, nil
	}

	// 如果是 PNG 格式，则转换为 JPEG
	var jpegBuffer bytes.Buffer
	err = jpeg.Encode(&jpegBuffer, img, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to encode image to JPEG: %v", err)
	}

	return jpegBuffer.Bytes(), nil
}

// 保存新的人脸样本到数据库和 Redis
func (l *FaceRecognitionLogic) saveNewFace(in *pb.FaceRecognitionRequest, faceFeatures *face.Face, hashKey string) (*pb.FaceRecognitionResponse, error) {
	// 人脸有效性判断 (大小必须大于50)
	if !l.isFaceValid(faceFeatures.Rectangle) {
		return nil, nil
	}

	// 保存人脸图片到本地
	faceImagePath, err := l.saveCroppedFaceToLocal(in.GetFace(), faceFeatures.Rectangle, in.GetUserId())
	if err != nil {
		return nil, err
	}

	// 保存到数据库
	storageFace, err := l.saveFaceToDatabase(in.GetUserId(), faceFeatures.Descriptor, faceImagePath)
	if err != nil {
		return nil, err
	}

	// 将新增数据写入 Redis
	err = l.appendFaceToRedisHash(hashKey, storageFace.ID, faceFeatures.Descriptor)
	if err != nil {
		return nil, fmt.Errorf("failed to append face to Redis: %v", err)
	}

	return &pb.FaceRecognitionResponse{
		FaceId: storageFace.ID,
	}, nil
}

// 加载数据库中的已有人脸
func (l *FaceRecognitionLogic) loadExistingFaces(userId string) ([]face.Descriptor, []int32, error) {
	if userId == "" {
		return nil, nil, fmt.Errorf("user ID is required")
	}
	storageFace := l.svcCtx.DB.ScaStorageFace
	existingFaces, err := storageFace.
		Select(storageFace.FaceVector, storageFace.ID).
		Where(storageFace.UserID.Eq(userId)).
		Find()
	if err != nil {
		return nil, nil, err
	}
	if len(existingFaces) == 0 {
		return nil, nil, nil
	}

	var samples []face.Descriptor
	var ids []int32
	// 使用并发处理每个数据
	for _, existingFace := range existingFaces {
		l.wg.Add(1)
		go func(faceData *model.ScaStorageFace) {
			defer l.wg.Done()

			var descriptor face.Descriptor
			if err = json.Unmarshal([]byte(faceData.FaceVector), &descriptor); err != nil {
				l.Errorf("failed to unmarshal face vector: %v", err)
				return
			}
			// 使用锁来保证并发访问时对切片的安全操作
			l.mu.Lock()
			samples = append(samples, descriptor)
			ids = append(ids, int32(faceData.ID))
			l.mu.Unlock()
		}(existingFace)
	}
	l.wg.Wait()
	return samples, ids, nil
}

const (
	minFaceWidth  = 50 // 最小允许的人脸宽度
	minFaceHeight = 50 // 最小允许的人脸高度
)

// 判断人脸是否有效
func (l *FaceRecognitionLogic) isFaceValid(rect image.Rectangle) bool {
	width := rect.Dx()
	height := rect.Dy()
	return width >= minFaceWidth && height >= minFaceHeight
}

// 保存人脸特征和路径到数据库
func (l *FaceRecognitionLogic) saveFaceToDatabase(userId string, descriptor face.Descriptor, faceImagePath string) (*model.ScaStorageFace, error) {
	jsonBytes, err := json.Marshal(descriptor)
	if err != nil {
		return nil, err
	}
	storageFace := model.ScaStorageFace{
		FaceVector:    string(jsonBytes),
		FaceImagePath: faceImagePath,
		UserID:        userId,
	}
	err = l.svcCtx.DB.ScaStorageFace.Create(&storageFace)
	if err != nil {
		return nil, err
	}
	return &storageFace, nil
}

func (l *FaceRecognitionLogic) saveCroppedFaceToLocal(faceImage []byte, rect image.Rectangle, userID string) (string, error) {
	objectKey := path.Join(
		userID,
		time.Now().Format("2006/01"), // 按年/月划分目录
		fmt.Sprintf("%s_%s.jpg", time.Now().Format("20060102150405"), kgo.SimpleUuid()),
	)

	// 解码图像
	img, _, err := image.Decode(bytes.NewReader(faceImage))
	if err != nil {
		return "", fmt.Errorf("image decode failed: %w", err)
	}

	// 获取图像边界
	imgBounds := img.Bounds()
	// 增加边距（比如 20 像素）
	margin := 20
	extendedRect := image.Rect(
		max(rect.Min.X-margin, imgBounds.Min.X), // 确保不超出左边界
		max(rect.Min.Y-margin, imgBounds.Min.Y), // 确保不超出上边界
		min(rect.Max.X+margin, imgBounds.Max.X), // 确保不超出右边界
		min(rect.Max.Y+margin, imgBounds.Max.Y), // 确保不超出下边界
	)
	// 裁剪图像
	croppedImage := img.(interface {
		SubImage(r image.Rectangle) image.Image
	}).SubImage(extendedRect)

	// 将图像编码为JPEG字节流
	var buf bytes.Buffer
	if err = jpeg.Encode(&buf, croppedImage, nil); err != nil {
		return "", fmt.Errorf("failed to encode image to JPEG: %w", err)
	}
	exists, err := l.svcCtx.MinioClient.BucketExists(l.ctx, constant.FaceBucketName)
	if err != nil || !exists {
		err = l.svcCtx.MinioClient.MakeBucket(l.ctx, constant.FaceBucketName, minio.MakeBucketOptions{Region: "us-east-1", ObjectLocking: true})
		if err != nil {
			logx.Errorf("Failed to create MinIO bucket: %v", err)
			return "", err
		}
	}

	// 上传到MinIO
	_, err = l.svcCtx.MinioClient.PutObject(
		l.ctx,
		constant.FaceBucketName,
		objectKey,
		bytes.NewReader(buf.Bytes()),
		int64(buf.Len()),
		minio.PutObjectOptions{
			ContentType: "image/jpeg",
		},
	)
	if err != nil {
		return "", fmt.Errorf("failed to upload image to MinIO: %w", err)
	}
	return objectKey, nil
}

// 从 Redis 的 Hash 中加载人脸数据
func (l *FaceRecognitionLogic) loadFacesFromRedisHash(hashKey string) ([]face.Descriptor, []int32, error) {
	// 从 Redis 获取 Hash 的所有字段和值
	data, err := l.svcCtx.RedisClient.HGetAll(l.ctx, hashKey).Result()
	if err != nil {
		return nil, nil, err
	}

	var samples []face.Descriptor
	var ids []int32
	for idStr, descriptorStr := range data {
		var descriptor face.Descriptor
		if err = json.Unmarshal([]byte(descriptorStr), &descriptor); err != nil {
			return nil, nil, err
		}

		// 转换 ID 为 int32
		id, err := parseInt32(idStr)
		if err != nil {
			return nil, nil, err
		}

		samples = append(samples, descriptor)
		ids = append(ids, id)
	}
	return samples, ids, nil
}

// 将人脸数据写入 Redis 的 Hash
func (l *FaceRecognitionLogic) cacheFacesToRedisHash(hashKey string, samples []face.Descriptor, ids []int32) error {
	// 开启事务
	pipe := l.svcCtx.RedisClient.Pipeline()

	for i := range samples {
		descriptorData, err := json.Marshal(samples[i])
		if err != nil {
			return err
		}

		// 使用 HSET 设置 Hash 字段和值
		pipe.HSet(l.ctx, hashKey, fmt.Sprintf("%d", ids[i]), descriptorData)
	}

	// 设置缓存过期时间
	pipe.Expire(l.ctx, hashKey, 3600*time.Second)

	_, err := pipe.Exec(l.ctx)
	return err
}

// 将新增的人脸数据追加到 Redis 的 Hash
func (l *FaceRecognitionLogic) appendFaceToRedisHash(hashKey string, id int64, descriptor face.Descriptor) error {
	descriptorData, err := json.Marshal(descriptor)
	if err != nil {
		return err
	}

	// 追加数据到 Hash
	err = l.svcCtx.RedisClient.HSet(l.ctx, hashKey, fmt.Sprintf("%d", id), descriptorData).Err()
	if err != nil {
		return err
	}
	// 检查是否已设置过期时间
	ttl, err := l.svcCtx.RedisClient.TTL(l.ctx, hashKey).Result()
	if err != nil {
		return err
	}

	// 如果未设置过期时间或已经过期，设置固定过期时间
	if ttl < 0 {
		err = l.svcCtx.RedisClient.Expire(l.ctx, hashKey, 3600*time.Second).Err()
		if err != nil {
			return err
		}
	}

	return nil
}

// 辅助函数：字符串转换为 int32
func parseInt32(s string) (int32, error) {
	var i int64
	var err error
	if i, err = strconv.ParseInt(s, 10, 32); err != nil {
		return 0, err
	}
	return int32(i), nil
}
