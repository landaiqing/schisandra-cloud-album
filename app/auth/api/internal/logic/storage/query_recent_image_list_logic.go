package storage

import (
	"context"
	"encoding/json"
	"errors"
	"schisandra-album-cloud-microservices/common/constant"
	"sync"
	"time"

	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type QueryRecentImageListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewQueryRecentImageListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *QueryRecentImageListLogic {
	return &QueryRecentImageListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *QueryRecentImageListLogic) QueryRecentImageList() (resp *types.RecentListResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	redisKeyPattern := constant.ImageRecentPrefix + uid + ":*"
	iter := l.svcCtx.RedisClient.Scan(l.ctx, 0, redisKeyPattern, 0).Iterator()
	var keys []string
	for iter.Next(l.ctx) {
		keys = append(keys, iter.Val())
	}
	if err := iter.Err(); err != nil {
		logx.Error(err)
		return nil, errors.New("scan recent file list failed")
	}

	if len(keys) == 0 {
		return &types.RecentListResponse{Records: []types.AllImageDetail{}}, nil
	}

	cmds, err := l.svcCtx.RedisClient.MGet(l.ctx, keys...).Result()
	if err != nil {
		logx.Error(err)
		return nil, errors.New("get recent file list failed")
	}

	var wg sync.WaitGroup
	groupedImages := sync.Map{}

	for _, cmd := range cmds {
		if cmd == nil {
			continue
		}
		val, ok := cmd.(string)
		if !ok {
			logx.Error("invalid value type")
			return nil, errors.New("invalid value type")
		}
		var imageMeta types.ImageMeta
		err = json.Unmarshal([]byte(val), &imageMeta)
		if err != nil {
			logx.Error(err)
			return nil, errors.New("unmarshal recent file list failed")
		}
		parse, err := time.Parse("2006-01-02 15:04:05", imageMeta.CreatedAt)
		if err != nil {
			logx.Error(err)
			return nil, errors.New("parse recent file list failed")
		}
		date := parse.Format("2006年1月2日 星期" + WeekdayMap[parse.Weekday()])
		// 使用LoadOrStore来检查并存储或者追加
		wg.Add(1)
		go func(date string, imageMeta types.ImageMeta) {
			defer wg.Done()
			value, loaded := groupedImages.LoadOrStore(date, []types.ImageMeta{imageMeta})
			if loaded {
				images := value.([]types.ImageMeta)
				images = append(images, imageMeta)
				groupedImages.Store(date, images)
			}
		}(date, imageMeta)
	}

	wg.Wait()

	var imageList []types.AllImageDetail
	groupedImages.Range(func(key, value interface{}) bool {
		imageList = append(imageList, types.AllImageDetail{
			Date: key.(string),
			List: value.([]types.ImageMeta),
		})
		return true
	})
	return &types.RecentListResponse{
		Records: imageList,
	}, nil
}
