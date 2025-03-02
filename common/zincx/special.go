package zincx

import (
	"fmt"
	"net/http"
)

// 创建索引（幂等操作，若存在则跳过）
func (zc *ZincClient) CreateFileUploadIndex(indexName string) error {
	exists, err := zc.IndexExists(indexName)
	if err != nil {
		return fmt.Errorf("检查索引失败: %w", err)
	}
	if exists {
		return nil // 索引已存在则跳过
	}

	// 定义完整的索引映射
	mapping := map[string]interface{}{
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				// 基础信息
				"storage_id": map[string]string{"type": "numeric"},
				"face_id":    map[string]string{"type": "numeric"},
				"file_name":  map[string]string{"type": "keyword"},
				"file_size":  map[string]string{"type": "numeric"},
				"uid":        map[string]string{"type": "keyword"},
				"file_path":  map[string]string{"type": "text"},
				"thumb_path": map[string]string{"type": "text"},
				"created_at": map[string]string{
					"type": "date"},

				// 文件元数据
				"provider":      map[string]string{"type": "keyword"},
				"bucket":        map[string]string{"type": "keyword"},
				"file_type":     map[string]string{"type": "keyword"},
				"is_anime":      map[string]string{"type": "boolean"},
				"tag_name":      map[string]string{"type": "keyword"},
				"landscape":     map[string]string{"type": "keyword"},
				"top_category":  map[string]string{"type": "keyword"},
				"is_screenshot": map[string]string{"type": "boolean"},

				// 媒体属性
				"width":      map[string]string{"type": "numeric"},
				"height":     map[string]string{"type": "numeric"},
				"thumb_w":    map[string]string{"type": "numeric"},
				"thumb_h":    map[string]string{"type": "numeric"},
				"thumb_size": map[string]string{"type": "numeric"},

				// 地理信息
				"longitude": map[string]string{"type": "numeric"},
				"latitude":  map[string]string{"type": "numeric"},
				"country":   map[string]string{"type": "keyword"},
				"province":  map[string]string{"type": "keyword"},
				"city":      map[string]string{"type": "keyword"},

				// 其他
				"album_id":   map[string]string{"type": "numeric"},
				"has_qrcode": map[string]string{"type": "boolean"},
			},
		},
	}

	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		SetHeader("Content-Type", "application/json").
		SetBody(mapping).
		Put(zc.BaseURL + "/api/index/" + indexName)

	if err != nil {
		return fmt.Errorf("创建索引请求失败: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("创建索引失败 (状态码 %d): %s", resp.StatusCode(), resp.String())
	}
	return nil
}
