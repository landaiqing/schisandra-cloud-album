package storage

import (
	"context"
	"errors"
	"fmt"
	"schisandra-album-cloud-microservices/common/constant"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewDeleteImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteImageLogic {
	return &DeleteImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *DeleteImageLogic) DeleteImage(req *types.DeleteImageRequest) (resp string, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return "", errors.New("user_id not found")
	}
	tx := l.svcCtx.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	storageInfo := tx.ScaStorageInfo
	info, err := storageInfo.Where(storageInfo.UserID.Eq(uid),
		storageInfo.ID.In(req.IDS...),
		storageInfo.Provider.Eq(req.Provider),
		storageInfo.Bucket.Eq(req.Bucket)).Delete()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if info.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("no image found")
	}
	storageThumb := tx.ScaStorageThumb
	resultInfo, err := storageThumb.Where(storageThumb.UserID.Eq(uid), storageThumb.InfoID.In(req.IDS...)).Delete()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if resultInfo.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("no thumb found")
	}
	storageExtra := tx.ScaStorageExtra
	resultExtra, err := storageExtra.Where(storageExtra.UserID.Eq(uid), storageExtra.InfoID.In(req.IDS...)).Delete()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if resultExtra.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("no extra found")
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	// 删除缓存
	keyPattern := fmt.Sprintf("%s%s:%s", constant.ImageCachePrefix, uid, "*")
	// 获取所有匹配的键
	keys, err := l.svcCtx.RedisClient.Keys(l.ctx, keyPattern).Result()
	if err != nil {
		logx.Errorf("获取缓存键 %s 失败: %v", keyPattern, err)
		return "", err
	}
	// 如果没有匹配的键，直接返回
	if len(keys) == 0 {
		logx.Infof("没有找到匹配的缓存键: %s", keyPattern)
		return "", nil
	}
	// 删除所有匹配的键
	if err := l.svcCtx.RedisClient.Del(l.ctx, keys...).Err(); err != nil {
		logx.Errorf("删除缓存键 %s 失败: %v", keyPattern, err)
		return "", err
	}
	return "success", nil
}
