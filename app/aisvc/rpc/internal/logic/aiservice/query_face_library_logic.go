package aiservicelogic

import (
	"context"
	"fmt"
	"net/url"
	"schisandra-album-cloud-microservices/app/aisvc/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"strconv"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/aisvc/rpc/internal/svc"
	"schisandra-album-cloud-microservices/app/aisvc/rpc/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryFaceLibraryLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
	wg sync.WaitGroup
	mu sync.Mutex
}

type FaceLibrary struct {
	ID        int64  `json:"id"`
	FaceImage []byte `json:"face_image"`
	FaceName  string `json:"face_name"`
}

func NewQueryFaceLibraryLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryFaceLibraryLogic {
	return &QueryFaceLibraryLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
		wg:     sync.WaitGroup{},
		mu:     sync.Mutex{},
	}
}

// QueryFaceLibrary queries the face library
func (l *QueryFaceLibraryLogic) QueryFaceLibrary(in *pb.QueryFaceLibraryRequest) (*pb.QueryFaceLibraryResponse, error) {
	if in.GetUserId() == "" {
		return nil, fmt.Errorf("user ID is required")
	}
	storageFace := l.svcCtx.DB.ScaStorageFace
	samples, err := storageFace.Select(
		storageFace.ID,
		storageFace.FaceVector,
		storageFace.FaceImagePath,
		storageFace.FaceName).
		Where(storageFace.UserID.Eq(in.GetUserId()), storageFace.FaceShow.Eq(in.GetType())).
		Find()
	if err != nil {
		return nil, fmt.Errorf("failed to query face library: %v", err)
	}
	if len(samples) == 0 {
		return nil, nil
	}
	faceLibrary := make([]*pb.FaceLibrary, len(samples))

	for i, sample := range samples {
		l.wg.Add(1)
		go func(i int, sample *model.ScaStorageFace) {
			defer l.wg.Done()
			redisKey := constant.FaceSamplePrefix + in.GetUserId() + ":" + strconv.FormatInt(sample.ID, 10)
			file, err := l.svcCtx.RedisClient.Get(l.ctx, redisKey).Result()
			if err == nil {
				l.mu.Lock()
				faceLibrary[i] = &pb.FaceLibrary{
					Id:        sample.ID,
					FaceName:  sample.FaceName,
					FaceImage: file,
				}
				l.mu.Unlock()
				return
			}
			reqParams := make(url.Values)
			presignedURL, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.FaceBucketName, sample.FaceImagePath, time.Hour*24, reqParams)

			err = l.svcCtx.RedisClient.Set(l.ctx, redisKey, presignedURL.String(), time.Hour*24).Err()
			if err != nil {
				return
			}
			l.mu.Lock()
			faceLibrary[i] = &pb.FaceLibrary{
				Id:        sample.ID,
				FaceName:  sample.FaceName,
				FaceImage: presignedURL.String(),
			}
			l.mu.Unlock()
		}(i, sample)
	}

	l.wg.Wait()

	return &pb.QueryFaceLibraryResponse{
		Faces: faceLibrary,
	}, nil
}
