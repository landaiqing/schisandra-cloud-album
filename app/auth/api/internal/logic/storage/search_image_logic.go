package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"schisandra-album-cloud-microservices/app/auth/api/internal/svc"
	"schisandra-album-cloud-microservices/app/auth/api/internal/types"
	"schisandra-album-cloud-microservices/app/auth/model/mysql/model"
	"schisandra-album-cloud-microservices/common/constant"
	"schisandra-album-cloud-microservices/common/encrypt"
	storageConfig "schisandra-album-cloud-microservices/common/storage/config"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SearchImageLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSearchImageLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchImageLogic {
	return &SearchImageLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SearchImageLogic) SearchImage(req *types.SearchImageRequest) (resp *types.SearchImageResponse, err error) {
	uid, ok := l.ctx.Value("user_id").(string)
	if !ok {
		return nil, errors.New("user_id not found")
	}

	baseQuery := map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{
					{"term": map[string]interface{}{"provider": req.Provider}},
					{"term": map[string]interface{}{"bucket": req.Bucket}},
					{"term": map[string]interface{}{"uid": uid}},
				},
			},
		},
	}
	switch req.Type {
	case "time":
		// 时间范围查询（示例："[2023-01-01,2023-12-31]"）
		start, end, err := parseTimeRange(req.Keyword)
		if err != nil {
			return nil, fmt.Errorf("时间解析失败: %w", err)
		}
		addTimeRangeQuery(baseQuery, start, end)
	case "person":
		// 人脸ID精确匹配
		faceID, err := strconv.ParseInt(req.Keyword, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("人脸ID格式错误: %w", err)
		}
		addFaceIDQuery(baseQuery, faceID)
	case "thing":
		// 标签和分类匹配
		addThingQuery(baseQuery, req.Keyword)
	case "picture":
		// 图片属性匹配
		addPictureQuery(baseQuery, req.Keyword)
	case "location":
		addLocationQuery(baseQuery, req.Keyword)
	default:
		return nil, errors.New("不支持的查询类型")
	}
	// 执行查询
	var target types.ZincFileInfo
	result, err := l.svcCtx.ZincClient.Search(constant.ZincIndexNameStorageInfo, baseQuery, target)
	if err != nil {
		return nil, fmt.Errorf("查询失败: %w", err)
	}

	// 加载用户oss配置信息
	cacheOssConfigKey := constant.UserOssConfigPrefix + uid + ":" + req.Provider
	ossConfig, err := l.getOssConfigFromCacheOrDb(cacheOssConfigKey, uid, req.Provider)
	if err != nil {
		return nil, err
	}

	service, err := l.svcCtx.StorageManager.GetStorage(uid, ossConfig)
	if err != nil {
		return nil, errors.New("get storage failed")
	}
	// 按日期分组处理
	groupedImages := sync.Map{}
	var wg sync.WaitGroup

	for _, hit := range result.Hits.Hits {
		wg.Add(1)
		go func(hit struct { // 明确传递 hit 结构
			ID     string      `json:"_id"`
			Source interface{} `json:"_source"`
			Score  float64     `json:"_score"`
		}) {
			defer wg.Done()

			// 类型断言转换
			source, err := convertToZincFileInfo(hit.Source)
			if err != nil {
				logx.Errorf("数据转换失败: %v | 原始数据: %+v", err, hit.Source)
				return
			}

			// 生成日期键（示例格式：2023年8月15日 星期二）
			weekday := WeekdayMap[source.CreatedAt.Weekday()]
			dateKey := source.CreatedAt.Format("2006年1月2日 星期" + weekday)

			// 生成访问链接
			thumbnailUrl, err := service.PresignedURL(l.ctx, ossConfig.BucketName, source.ThumbPath, 15*time.Minute)
			if err != nil {
				logx.Errorf("生成缩略图链接失败: %v", err)
				return
			}

			fileUrl, err := service.PresignedURL(l.ctx, ossConfig.BucketName, source.FilePath, 15*time.Minute)
			if err != nil {
				logx.Errorf("生成文件链接失败: %v", err)
				return
			}

			// 构建元数据
			meta := types.ImageMeta{
				ID:        source.StorageId,
				FileName:  source.FileName,
				URL:       fileUrl,
				Width:     source.ThumbW,
				Height:    source.ThumbH,
				CreatedAt: source.CreatedAt.Format("2006-01-02 15:04:05"),
				Thumbnail: thumbnailUrl,
			}

			// 线程安全写入 Map
			value, _ := groupedImages.LoadOrStore(dateKey, []types.ImageMeta{})
			images := value.([]types.ImageMeta)
			images = append(images, meta)
			groupedImages.Store(dateKey, images)
		}(struct {
			ID     string      `json:"_id"`
			Source interface{} `json:"_source"`
			Score  float64     `json:"_score"`
		}(hit)) // 将 hit 作为参数传递
	}
	wg.Wait()

	// 转换分组结果
	var records []types.AllImageDetail
	groupedImages.Range(func(key, value interface{}) bool {
		records = append(records, types.AllImageDetail{
			Date: key.(string),
			List: value.([]types.ImageMeta),
		})
		return true
	})
	// 按日期降序排序
	sort.Slice(records, func(i, j int) bool {
		ti, _ := time.Parse("2006年1月2日 星期一", records[i].Date)
		tj, _ := time.Parse("2006年1月2日 星期一", records[j].Date)
		return ti.After(tj)
	})
	return &types.SearchImageResponse{
		Records: records,
	}, nil

}

// 时间范围解析（支持日期和时间戳）
func parseTimeRange(input string) (int64, int64, error) {
	input = strings.Trim(input, "[]")
	parts := strings.Split(input, ",")
	if len(parts) != 2 {
		return 0, 0, errors.New("时间格式错误")
	}

	parseTime := func(s string) (int64, error) {
		// 尝试解析为日期格式
		if t, err := time.Parse("2006-01-02", strings.TrimSpace(s)); err == nil {
			return t.Unix(), nil
		}
		// 尝试解析为时间戳
		if ts, err := strconv.ParseInt(s, 10, 64); err == nil {
			return ts, nil
		}
		return 0, errors.New("无效时间格式")
	}

	start, err := parseTime(parts[0])
	if err != nil {
		return 0, 0, err
	}
	end, err := parseTime(parts[1])
	if err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func addTimeRangeQuery(query map[string]interface{}, start, end int64) {
	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"]
	timeQuery := map[string]interface{}{
		"range": map[string]interface{}{
			"created_at": map[string]interface{}{ // 改为使用 created_at 字段
				"gte": start,
				"lte": end,
			},
		},
	}
	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(must.([]map[string]interface{}), timeQuery)
}

// 修改后的标签查询
func addThingQuery(query map[string]interface{}, keyword string) {
	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"]
	tagQuery := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  keyword,
			"fields": []string{"tag_name", "top_category"}, // 使用新字段名
		},
	}
	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(must.([]map[string]interface{}), tagQuery)
}

// 修改后的文件类型查询
// 修改后的 picture 类型查询 (同时搜索文件名和文件类型)
func addPictureQuery(query map[string]interface{}, keyword string) {
	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"]

	pictureQuery := map[string]interface{}{
		"bool": map[string]interface{}{
			"should": []map[string]interface{}{
				{
					"wildcard": map[string]interface{}{
						"tag_name": "*" + strings.ToLower(keyword) + "*",
					},
				},
				{
					"term": map[string]interface{}{
						"top_category": strings.ToLower(keyword),
					},
				},
			},
			"minimum_should_match": 1,
		},
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
		must.([]map[string]interface{}),
		pictureQuery,
	)
}

// 添加人脸ID查询
func addFaceIDQuery(query map[string]interface{}, faceID int64) {
	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"].([]map[string]interface{})
	idQuery := map[string]interface{}{
		"term": map[string]interface{}{
			"face_id": faceID,
		},
	}
	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(must, idQuery)
}

func addLocationQuery(query map[string]interface{}, keyword string) {
	must := query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"]

	locationQuery := map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":  keyword,
			"fields": []string{"country", "province", "city"},
			"type":   "best_fields", // 优先匹配最多字段
		},
	}

	query["query"].(map[string]interface{})["bool"].(map[string]interface{})["must"] = append(
		must.([]map[string]interface{}),
		locationQuery,
	)
}

// ZincSearch 响应结构
type ZincSearchResult struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string                  `json:"_id"`
			Source types.FileUploadMessage `json:"_source"`
			Score  float64                 `json:"_score"`
		} `json:"hits"`
	} `json:"hits"`
}

// 提取解密操作为函数
func (l *SearchImageLogic) decryptConfig(config *model.ScaStorageConfig) (*storageConfig.StorageConfig, error) {
	accessKey, err := encrypt.Decrypt(config.AccessKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt access key failed")
	}
	secretKey, err := encrypt.Decrypt(config.SecretKey, l.svcCtx.Config.Encrypt.Key)
	if err != nil {
		return nil, errors.New("decrypt secret key failed")
	}
	return &storageConfig.StorageConfig{
		Provider:   config.Provider,
		Endpoint:   config.Endpoint,
		AccessKey:  accessKey,
		SecretKey:  secretKey,
		BucketName: config.Bucket,
		Region:     config.Region,
	}, nil
}

// 从缓存或数据库中获取 OSS 配置
func (l *SearchImageLogic) getOssConfigFromCacheOrDb(cacheKey, uid, provider string) (*storageConfig.StorageConfig, error) {
	result, err := l.svcCtx.RedisClient.Get(l.ctx, cacheKey).Result()
	if err != nil && !errors.Is(err, redis.Nil) {
		return nil, errors.New("get oss config failed")
	}

	var ossConfig *storageConfig.StorageConfig
	if result != "" {
		var redisOssConfig model.ScaStorageConfig
		if err = json.Unmarshal([]byte(result), &redisOssConfig); err != nil {
			return nil, errors.New("unmarshal oss config failed")
		}
		return l.decryptConfig(&redisOssConfig)
	}

	// 缓存未命中，从数据库中加载
	scaOssConfig := l.svcCtx.DB.ScaStorageConfig
	dbOssConfig, err := scaOssConfig.Where(scaOssConfig.UserID.Eq(uid), scaOssConfig.Provider.Eq(provider)).First()
	if err != nil {
		return nil, err
	}

	// 缓存数据库配置
	ossConfig, err = l.decryptConfig(dbOssConfig)
	if err != nil {
		return nil, err
	}
	marshalData, err := json.Marshal(dbOssConfig)
	if err != nil {
		return nil, errors.New("marshal oss config failed")
	}
	err = l.svcCtx.RedisClient.Set(l.ctx, cacheKey, marshalData, 0).Err()
	if err != nil {
		return nil, errors.New("set oss config failed")
	}

	return ossConfig, nil
}

// 新的数据转换函数
func convertToZincFileInfo(source interface{}) (types.ZincFileInfo, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return types.ZincFileInfo{}, fmt.Errorf("序列化失败: %w", err)
	}

	var info types.ZincFileInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return types.ZincFileInfo{}, fmt.Errorf("反序列化失败: %w", err)
	}
	return info, nil
}
