package share

import (
	"context"
	"errors"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type ListShareRecordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewListShareRecordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *ListShareRecordLogic {
	return &ListShareRecordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *ListShareRecordLogic) ListShareRecord(req *types.ShareRecordListRequest) (resp *types.ShareRecordListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageShare := l.svcCtx.DB.ScaStorageShare
	storageAlbum := l.svcCtx.DB.ScaStorageAlbum
	var recordList []types.ShareRecord
	query := storageShare.
		Select(storageShare.ID,
			storageShare.InviteCode,
			storageShare.VisitLimit,
			storageShare.AccessPassword,
			storageShare.ValidityPeriod,
			storageShare.CreatedAt,
			storageAlbum.CoverImage).
		LeftJoin(storageAlbum, storageShare.AlbumID.EqCol(storageAlbum.ID)).
		Where(storageShare.UserID.Eq(uid)).
		Order(storageShare.CreatedAt.Desc())

	if len(req.DateRange) == 2 {
		startDate, errStart := time.Parse("2006-01-02", req.DateRange[0])
		endDate, errEnd := time.Parse("2006-01-02", req.DateRange[1])
		if errStart != nil || errEnd != nil {
			return nil, errors.New("invalid date format")
		}
		// Ensure endDate is inclusive by adding 24 hours
		endDate = endDate.AddDate(0, 0, 1)
		query = query.Where(storageShare.CreatedAt.Between(startDate, endDate))
	}
	err = query.Scan(&recordList)
	if err != nil {
		return nil, err
	}

	resp = &types.ShareRecordListResponse{
		Records: recordList,
	}
	return resp, nil
}
