package share

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryShareOverviewLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryShareOverviewLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryShareOverviewLogic {
	return &QueryShareOverviewLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryShareOverviewLogic) QueryShareOverview() (resp *types.ShareOverviewResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageShare := l.svcCtx.DB.ScaStorageShare
	shareVisit := l.svcCtx.DB.ScaStorageShareVisit
	// 统计所有数据
	var totalResult struct {
		TotalCount int64
		TotalViews int64
		TotalUsers int64
	}
	err = storageShare.Select(
		storageShare.ID.Count().As("total_count"),
		shareVisit.Views.Sum().As("total_views"),
		shareVisit.UserID.Distinct().Count().As("total_users"),
	).
		Join(shareVisit, storageShare.ID.EqCol(shareVisit.ShareID)).
		Where(storageShare.UserID.Eq(uid)).
		Scan(&totalResult)
	if err != nil {
		return nil, err
	}
	// 统计当天数据
	var dailyResult struct {
		DailyCount int64
		DailyViews int64
		DailyUsers int64
	}
	err = storageShare.Select(
		storageShare.ID.Count().As("daily_count"),
		shareVisit.Views.Sum().As("daily_views"),
		shareVisit.UserID.Distinct().Count().As("daily_users"),
	).
		Join(shareVisit, storageShare.ID.EqCol(shareVisit.ShareID)).
		Where(storageShare.UserID.Eq(uid),
			shareVisit.CreatedAt.Gte(time.Now().Truncate(24*time.Hour))).
		Scan(&dailyResult)
	if err != nil {
		return nil, err
	}
	// 合并结果到 ShareOverviewResponse
	response := types.ShareOverviewResponse{
		VisitCount:        totalResult.TotalViews, // 总访问量
		VisitCountToday:   dailyResult.DailyViews, // 当天访问量
		ViewerCount:       totalResult.TotalUsers, // 总独立用户数
		ViewerCountToday:  dailyResult.DailyUsers, // 当天独立用户数
		PublishCount:      totalResult.TotalCount, // 总发布量
		PublishCountToday: dailyResult.DailyCount, // 当天发布量
	}

	return &response, nil
}
