package storage

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/common/constant"
	"time"

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

func (l *QueryLocationImageListLogic) QueryLocationImageList(req *types.LocationListRequest) (resp *types.LocationListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageLocation := l.svcCtx.DB.ScaStorageLocation
	storageInfo := l.svcCtx.DB.ScaStorageInfo

	var locations []types.LocationInfo
	err = storageLocation.Select(
		storageLocation.ID,
		storageLocation.Country,
		storageLocation.City,
		storageLocation.Province,
		storageLocation.CoverImage,
		storageInfo.ID.Count().As("total")).
		LeftJoin(storageInfo, storageInfo.LocationID.EqCol(storageLocation.ID)).
		Where(storageLocation.UserID.Eq(uid),
			storageInfo.Provider.Eq(req.Provider),
			storageInfo.Bucket.Eq(req.Bucket)).
		Order(storageLocation.CreatedAt.Desc()).
		Group(storageLocation.ID).
		Scan(&locations)
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
		reqParams := make(url.Values)
		presignedUrl, err := l.svcCtx.MinioClient.PresignedGetObject(l.ctx, constant.ThumbnailBucketName, loc.CoverImage, 15*time.Minute, reqParams)
		if err != nil {
			return nil, errors.New("get presigned url failed")
		}
		locationMeta := types.LocationMeta{
			ID:         loc.ID,
			City:       city,
			Total:      loc.Total,
			CoverImage: presignedUrl.String(),
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
