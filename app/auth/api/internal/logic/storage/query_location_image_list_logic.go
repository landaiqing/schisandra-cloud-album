package storage

import (
	"context"
	"errors"
	"fmt"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryLocationImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryLocationImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryLocationImageListLogic {
	return &QueryLocationImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryLocationImageListLogic) QueryLocationImageList() (resp *types.LocationListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageLocation := l.svcCtx.DB.ScaStorageLocation

	locations, err := storageLocation.Select(
		storageLocation.ID,
		storageLocation.Country,
		storageLocation.City,
		storageLocation.Province,
		storageLocation.Total).Where(storageLocation.UserID.Eq(uid)).
		Order(storageLocation.CreatedAt.Desc()).Find()
	if err != nil {
		return nil, err
	}
	locationMap := make(map[string][]types.LocationMeta)

	for _, loc := range locations {
		var locationKey string
		if loc.Province == "" {
			locationKey = loc.Country
		} else {
			locationKey = fmt.Sprintf("%s %s", loc.Country, loc.Province)
		}

		city := loc.City
		if city == "" {
			city = loc.Country
		}
		locationMeta := types.LocationMeta{
			ID:    loc.ID,
			City:  city,
			Total: loc.Total,
		}
		locationMap[locationKey] = append(locationMap[locationKey], locationMeta)
	}

	var locationListData []types.LocationListData

	for location, list := range locationMap {
		locationListData = append(locationListData, types.LocationListData{
			Location: location,
			List:     list,
		})
	}

	return &types.LocationListResponse{Records: locationListData}, nil
}
