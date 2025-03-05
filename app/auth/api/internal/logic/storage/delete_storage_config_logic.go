package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/common/constant"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteStorageConfigLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteStorageConfigLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteStorageConfigLogic {
	return &DeleteStorageConfigLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteStorageConfigLogic) DeleteStorageConfig(req *types.DeleteStorageConfigRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	storageConfig := l.svcCtx.DB.ScaStorageConfig
	info, err := storageConfig.Where(storageConfig.ID.Eq(req.ID), storageConfig.UserID.Eq(uid),
		storageConfig.Provider.Eq(req.Provider), storageConfig.Bucket.Eq(req.Bucket)).Delete()
	if err != nil {
		return "", err
	}
	if info.RowsAffected == 0 {
		return "", errors.New("storage config not found")
	}
	cacheOssConfigKey := constant.UserOssConfigPrefix + uid + ":" + req.Provider
	err = l.svcCtx.RedisClient.Del(l.ctx, cacheOssConfigKey).Err()
	if err != nil {
		return "", err
	}
	return "success", nil
}
