package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListUserStorageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListUserStorageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListUserStorageLogic {
	return &ListUserStorageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListUserStorageLogic) ListUserStorage() (resp *types.StorageConfigListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageConfig := l.svcCtx.DB.ScaStorageConfig

	storageConfigList, err := storageConfig.Select(
		storageConfig.ID,
		storageConfig.Provider,
		storageConfig.Bucket,
		storageConfig.Capacity,
		storageConfig.CreatedAt,
		storageConfig.Region,
		storageConfig.Endpoint).
		Where(storageConfig.UserID.Eq(uid)).
		Order(storageConfig.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}
	var result []types.StorageConfigMeta
	for _, info := range storageConfigList {
		result = append(result, types.StorageConfigMeta{
			ID:        info.ID,
			Provider:  info.Provider,
			Endpoint:  info.Endpoint,
			Bucket:    info.Bucket,
			Capacity:  info.Capacity,
			Region:    info.Region,
			CreatedAt: info.CreatedAt.Format("2006-01-02"),
		})
	}
	return &types.StorageConfigListResponse{
		Records: result,
	}, nil
}
