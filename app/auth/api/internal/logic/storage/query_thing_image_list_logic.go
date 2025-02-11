package storage

import (
	"context"
	"errors"
	"sync"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryThingImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryThingImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryThingImageListLogic {
	return &QueryThingImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryThingImageListLogic) QueryThingImageList() (resp *types.ThingListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}
	storageInfo := l.svcCtx.DB.ScaStorageInfo
	storageInfos, err := storageInfo.Select(
		storageInfo.ID,
		storageInfo.Category,
		storageInfo.Tags,
		storageInfo.CreatedAt).
		Where(storageInfo.UserID.Eq(uid),
			storageInfo.Category.IsNotNull(),
			storageInfo.Tags.IsNotNull()).
		Order(storageInfo.CreatedAt.Desc()).
		Find()
	if err != nil {
		return nil, err
	}

	categoryMap := sync.Map{}
	tagCountMap := sync.Map{}

	for _, info := range storageInfos {
		tagKey := info.Category + "::" + info.Tags
		if _, exists := tagCountMap.Load(tagKey); !exists {
			tagCountMap.Store(tagKey, int64(0))
			categoryEntry, _ := categoryMap.LoadOrStore(info.Category, &sync.Map{})
			tagMap := categoryEntry.(*sync.Map)
			tagMap.Store(info.Tags, types.ThingMeta{
				TagName:   info.Tags,
				CreatedAt: info.CreatedAt.Format("2006-01-02 15:04:05"),
			})
		}
		tagCount, _ := tagCountMap.Load(tagKey)
		tagCountMap.Store(tagKey, tagCount.(int64)+1)
	}

	var thingListData []types.ThingListData
	categoryMap.Range(func(category, tagData interface{}) bool {
		var metas []types.ThingMeta
		tagData.(*sync.Map).Range(func(tag, item interface{}) bool {
			tagKey := category.(string) + "::" + tag.(string)
			tagCount, _ := tagCountMap.Load(tagKey)
			meta := item.(types.ThingMeta)
			meta.TagCount = tagCount.(int64)
			metas = append(metas, meta)
			return true
		})
		thingListData = append(thingListData, types.ThingListData{
			Category: category.(string),
			List:     metas,
		})
		return true
	})

	return &types.ThingListResponse{Records: thingListData}, nil
}
