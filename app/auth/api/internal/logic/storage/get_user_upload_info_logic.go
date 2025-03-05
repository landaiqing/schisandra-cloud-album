package storage

import (
	"context"
	"errors"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserUploadInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserUploadInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserUploadInfoLogic {
	return &GetUserUploadInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserUploadInfoLogic) GetUserUploadInfo() (resp *types.UserUploadInfoResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	// 图片数
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	var imageResult struct {
		ImageCount    int64 `json:"image_count"`
		FileSizeCount int64 `json:"file_size_count"`
	}
	err = storageInfo.Select(storageInfo.ID.Count().As("image_count"),
		storageInfo.FileSize.Sum().As("file_size_count")).Where(storageInfo.UserID.Eq(uid)).Scan(&imageResult)
	if err != nil {
		return nil, err
	}
	// 分享数
	storageShare := l.svcCtx.DB.ScaStorageShare
	var shareCount int64
	err = storageShare.Select(storageShare.ID.Count().As("share_count")).Where(storageShare.UserID.Eq(uid)).Scan(&shareCount)
	if err != nil {
		return nil, err
	}
	// 今日上传数
	var todayResult struct {
		TodayUploadCount   int64 `json:"today_upload_count"`
		TodayFileSizeCount int64 `json:"today_file_size_count"`
	}
	err = storageInfo.Select(
		storageInfo.ID.Count().As("today_upload_count"),
		storageInfo.FileSize.Sum().As("today_file_size_count")).Where(storageInfo.UserID.Eq(uid),
		storageInfo.CreatedAt.Gte(time.Now().Truncate(24*time.Hour))).Scan(&todayResult)
	if err != nil {
		return nil, err
	}

	// 今日分享数
	var todayShareCount int64
	err = storageShare.Select(storageShare.ID.Count().As("today_share_count")).Where(storageShare.UserID.Eq(uid),
		storageShare.CreatedAt.Gte(time.Now().Truncate(24*time.Hour))).Scan(&todayShareCount)
	if err != nil {
		return nil, err
	}

	// 热力图
	heatmap := make([]types.HeatmapMeta, 0)
	err = storageInfo.Select(
		storageInfo.CreatedAt.Date().As("date"),
		storageInfo.ID.Count().As("count"),
	).
		Where(storageInfo.UserID.Eq(uid)).
		Group(storageInfo.CreatedAt.Date()).
		Order(storageInfo.CreatedAt.Date().Desc()).
		Scan(&heatmap)
	if err != nil {
		return nil, err
	}
	resp = &types.UserUploadInfoResponse{
		ImageCount:         imageResult.ImageCount,
		TodayUploadCount:   todayResult.TodayUploadCount,
		ShareCount:         shareCount,
		TodayShareCount:    todayShareCount,
		FileSizeCount:      imageResult.FileSizeCount,
		TodayFileSizeCount: todayResult.TodayFileSizeCount,
		Heatmap:            heatmap,
	}

	return resp, nil
}
