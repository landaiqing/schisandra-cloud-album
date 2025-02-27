package storage

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/ccpwcn/kgo"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"strconv"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ShareAlbumLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewShareAlbumLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ShareAlbumLogic {
	return &ShareAlbumLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ShareAlbumLogic) ShareAlbum(req *types.ShareAlbumRequest) (resp string, err error) {
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

	storageAlbum := tx.ScaStorageAlbum
	info, err := storageAlbum.Where(storageAlbum.ID.Eq(req.ID), storageAlbum.UserID.Eq(uid)).
		Update(storageAlbum.AlbumType, constant.AlbumTypeShared)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	if info.RowsAffected == 0 {
		tx.Rollback()
		return "", errors.New("album not found")
	}
	// 更新图片信息
	storageInfo := tx.ScaStorageInfo
	_, err = storageInfo.Where(storageInfo.AlbumID.Eq(req.ID), storageInfo.UserID.Eq(uid)).
		Update(storageInfo.Type, constant.ImageTypeShared)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	// 查询图片数量
	var imageCount int64
	err = storageInfo.Select(
		storageInfo.ID.Count().As("image_count")).
		Where(storageInfo.AlbumID.Eq(req.ID), storageInfo.UserID.Eq(uid)).
		Group(storageInfo.AlbumID).Scan(&imageCount)
	if err != nil {
		tx.Rollback()
		return "", err
	}

	duration, err := strconv.Atoi(req.ExpireDate)
	if err != nil {
		return "", errors.New("invalid expire date")
	}
	expiryTime := l.GenerateExpiryTime(time.Now(), duration)
	storageShare := tx.ScaStorageShare
	storageShareInfo := &model.ScaStorageShare{
		UserID:         uid,
		AlbumID:        req.ID,
		InviteCode:     kgo.SimpleUuid(),
		ValidityPeriod: int64(duration),
		ExpireTime:     expiryTime,
		AccessPassword: req.AccessPassword,
		VisitLimit:     req.AccessLimit,
		ImageCount:     imageCount,
		Status:         0,
	}
	err = storageShare.Create(storageShareInfo)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	marshal, err := json.Marshal(storageShareInfo)
	if err != nil {
		tx.Rollback()
		return "", err
	}
	cacheKey := constant.ImageSharePrefix + storageShareInfo.InviteCode
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, marshal, time.Duration(duration)*time.Hour*24).Err()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", err
	}

	return "success", nil
}

// GenerateExpiryTime 函数接受当前时间和有效期（天为单位），返回过期时间
func (l *ShareAlbumLogic) GenerateExpiryTime(currentTime time.Time, durationInDays int) time.Time {
	// 创建一个持续时间对象
	duration := time.Duration(durationInDays) * 24 * time.Hour
	// 将当前时间加上持续时间，得到过期时间
	expiryTime := currentTime.Add(duration)
	return expiryTime
}
