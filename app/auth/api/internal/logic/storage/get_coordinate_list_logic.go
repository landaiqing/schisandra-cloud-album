package storage

import (
	"context"
	"errors"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCoordinateListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCoordinateListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCoordinateListLogic {
	return &GetCoordinateListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCoordinateListLogic) GetCoordinateList() (resp *types.CoordinateListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageLocation := l.svcCtx.DB.ScaStorageLocation
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	var records []types.CoordinateMeta
	err = storageLocation.Select(
		storageLocation.ID,
		storageLocation.Longitude,
		storageLocation.Latitude,
		storageLocation.Country,
		storageLocation.Province,
		storageLocation.City,
		storageInfo.ID.Count().As("image_count"),
	).Join(
		storageInfo,
		storageLocation.ID.EqCol(storageInfo.LocationID),
	).Where(storageLocation.UserID.Eq(uid),
		storageInfo.UserID.Eq(uid),
	).
		Group(storageLocation.ID).
		Scan(&records)
	if err != nil {
		return nil, err
	}
	return &types.CoordinateListResponse{
		Records: records,
	}, nil
}
