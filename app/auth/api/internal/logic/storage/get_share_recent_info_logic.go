package storage

import (
	"context"
	"errors"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetShareRecentInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetShareRecentInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetShareRecentInfoLogic {
	return &GetShareRecentInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetShareRecentInfoLogic) GetShareRecentInfo() (resp *types.ShareRecentInfoResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	// 生成最近7天的日期列表（包含今天）
	now := time.Now().Truncate(24 * time.Hour)
	dates := make([]string, 7)
	for i := 0; i < 7; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		dates[6-i] = date // 保证日期顺序从旧到新
	}

	// 计算查询时间范围
	startDate := now.AddDate(0, 0, -6)
	endDate := now
	// 查询每日发布次数
	var publishStats []struct {
		Date  string
		Count int64
	}
	storageShare := l.svcCtx.DB.ScaStorageShare
	err = storageShare.
		Select(storageShare.CreatedAt.Date().As("date"),
			storageShare.ID.Count().As("count")).
		Where(storageShare.UserID.Eq(uid)).
		Where(storageShare.CreatedAt.Between(startDate, endDate)).
		Group(storageShare.CreatedAt.Date()).
		Scan(&publishStats)
	if err != nil {
		return nil, err
	}
	// 查询每日访问数据
	var visitStats []struct {
		Date         string
		VisitCount   int64
		VisitorCount int64
	}
	shareVisit := l.svcCtx.DB.ScaStorageShareVisit
	err = shareVisit.Select(
		shareVisit.CreatedAt.Date().As("date"),
		shareVisit.ID.Count().As("visit_count"),
		shareVisit.UserID.Distinct().Count().As("visitor_count")).
		Join(storageShare, shareVisit.ShareID.EqCol(storageShare.ID)).
		Where(storageShare.UserID.Eq(uid),
			shareVisit.CreatedAt.Between(startDate, endDate)).
		Group(shareVisit.CreatedAt.Date()).
		Scan(&visitStats)
	if err != nil {
		return nil, err
	}
	// 初始化结果映射
	resultMap := make(map[string]*types.ShareRecentMeta)
	for _, date := range dates {
		resultMap[date] = &types.ShareRecentMeta{
			Date:         date,
			VisitCount:   0,
			VisitorCount: 0,
			PublishCount: 0,
		}
	}

	// 填充发布数据
	for _, stat := range publishStats {
		if meta, exists := resultMap[stat.Date]; exists {
			meta.PublishCount = stat.Count
		}
	}

	// 填充访问数据
	for _, stat := range visitStats {
		if meta, exists := resultMap[stat.Date]; exists {
			meta.VisitCount = stat.VisitCount
			meta.VisitorCount = stat.VisitorCount
		}
	}

	// 构建有序结果
	records := make([]types.ShareRecentMeta, 0, 7)
	for _, date := range dates {
		records = append(records, *resultMap[date])
	}

	return &types.ShareRecentInfoResponse{Records: records}, nil
}
