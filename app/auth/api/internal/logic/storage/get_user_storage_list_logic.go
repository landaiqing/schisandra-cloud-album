package storage

import (
	"context"
	"errors"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserStorageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserStorageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserStorageListLogic {
	return &GetUserStorageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// providerNameMap 存储商映射表
var providerNameMap = map[string]string{
	"ali":     "阿里云OSS",
	"tencent": "腾讯云COS",
	"aws":     "Amazon S3",
	"qiniu":   "七牛云",
	"huawei":  "华为云OBS",
}

func (l *GetUserStorageListLogic) GetUserStorageList() (resp *types.StorageListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageConfig := l.svcCtx.DB.ScaStorageConfig
	storageConfigs, err := storageConfig.Select(
		storageConfig.Provider,
		storageConfig.Bucket).
		Where(
			storageConfig.UserID.Eq(uid)).Find()
	if err != nil {
		return nil, err
	}
	// 使用 map 组织数据
	providerMap := make(map[string][]types.StorageMeta)

	for _, config := range storageConfigs {
		providerMap[config.Provider] = append(providerMap[config.Provider], types.StorageMeta{
			Value: config.Bucket,
			Name:  config.Bucket,
		})
	}
	// 组装返回结构
	var records []types.StroageNode
	for provider, buckets := range providerMap {
		records = append(records, types.StroageNode{
			Value:    provider,
			Name:     l.getProviderName(provider),
			Children: buckets,
		})
	}
	// 返回数据
	return &types.StorageListResponse{
		Records: records,
	}, nil
}

// getProviderName 获取存储商的中文名称
func (l *GetUserStorageListLogic) getProviderName(provider string) string {
	if name, exists := providerNameMap[provider]; exists {
		return name
	}
	return provider
}
