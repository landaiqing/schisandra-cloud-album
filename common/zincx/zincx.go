package zincx

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"net/http"
)

type ZincClient struct {
	Client   *resty.Client
	BaseURL  string
	Username string
	Password string
}

func NewZincClient(BaseURL, Username, Password string) *ZincClient {
	Client := resty.New().
		SetDebug(false).
		SetDisableWarn(true)
	return &ZincClient{
		Client:   Client,
		BaseURL:  BaseURL,
		Username: Username,
		Password: Password,
	}
}

// 检查索引是否存在 (内部方法)
func (zc *ZincClient) IndexExists(indexName string) (bool, error) {
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		Head(zc.BaseURL + "/api/index/" + indexName)

	if err != nil {
		return false, err
	}

	return resp.StatusCode() == http.StatusOK, nil
}

type IndexMapping struct {
	Mappings struct {
		Properties map[string]interface{} `json:"properties"`
	} `json:"mappings"`
}

func (zc *ZincClient) CreateIndex(indexName string, mapping *IndexMapping) error {
	url := fmt.Sprintf("%s/api/index/%s", zc.BaseURL, indexName)
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		SetHeader("Content-Type", "application/json").
		SetBody(mapping).
		Put(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("创建索引失败: %s", resp.String())
	}
	return nil
}

func (zc *ZincClient) IndexDocument(indexName, documentID string, doc interface{}) error {
	url := fmt.Sprintf("%s/api/%s/_doc/%s", zc.BaseURL, indexName, documentID)
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		SetHeader("Content-Type", "application/json").
		SetBody(doc).
		Put(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("插入文档失败: %s", resp.String())
	}
	return nil
}

type SearchRequest struct {
	Query struct {
		Match map[string]interface{} `json:"match"`
	} `json:"query"`
}

// 泛型响应结构
type ZincSearchResponse[T any] struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			ID     string  `json:"_id"`
			Source T       `json:"_source"`
			Score  float64 `json:"_score"`
		} `json:"hits"`
	} `json:"hits"`
}

// 修改 Search 方法签名
func (zc *ZincClient) Search(indexName string, query interface{}, resultType interface{}) (*ZincSearchResponse[interface{}], error) {
	url := fmt.Sprintf("%s/api/%s/_search", zc.BaseURL, indexName)
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		SetHeader("Content-Type", "application/json").
		SetBody(query).
		SetResult(&ZincSearchResponse[interface{}]{}).
		Post(url)

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != 200 {
		return nil, fmt.Errorf("搜索失败: %s", resp.String())
	}

	// 手动反序列化以支持动态类型
	var rawResponse ZincSearchResponse[interface{}]
	if err := json.Unmarshal(resp.Body(), &rawResponse); err != nil {
		return nil, err
	}

	// 将 Source 转换为目标类型
	for i := range rawResponse.Hits.Hits {
		data, _ := json.Marshal(rawResponse.Hits.Hits[i].Source)
		_ = json.Unmarshal(data, &resultType)
		rawResponse.Hits.Hits[i].Source = resultType
	}

	return &rawResponse, nil
}

type BulkRequest struct {
	Index  string      `json:"index"`
	ID     string      `json:"id"`
	Source interface{} `json:"source"`
}

func (zc *ZincClient) BulkIndex(indexName string, docs []BulkRequest) error {
	url := fmt.Sprintf("%s/api/_bulk", zc.BaseURL)
	body := ""
	for _, doc := range docs {
		action := fmt.Sprintf(`{ "index": { "_index": "%s", "_id": "%s" } }`, indexName, doc.ID)
		source, _ := json.Marshal(doc.Source)
		body += action + "\n" + string(source) + "\n"
	}

	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		SetHeader("Content-Type", "application/json").
		SetBody(body).
		Post(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("批量操作失败: %s", resp.String())
	}
	return nil
}

// 删除文档
func (zc *ZincClient) DeleteDocument(indexName, documentID string) error {
	url := fmt.Sprintf("%s/api/%s/_doc/%s", zc.BaseURL, indexName, documentID)
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		Delete(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("删除文档失败: %s", resp.String())
	}
	return nil
}

// 删除索引
func (zc *ZincClient) DeleteIndex(indexName string) error {
	url := fmt.Sprintf("%s/api/index/%s", zc.BaseURL, indexName)
	resp, err := zc.Client.R().
		SetBasicAuth(zc.Username, zc.Password).
		Delete(url)

	if err != nil {
		return err
	}

	if resp.StatusCode() != 200 {
		return fmt.Errorf("删除索引失败: %s", resp.String())
	}
	return nil
}
