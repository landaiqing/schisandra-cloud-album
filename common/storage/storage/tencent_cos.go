package storage

import (
	"github.com/tencentyun/cos-go-sdk-v5"
	"schisandra-album-cloud-microservices/common/storage/events"

	"schisandra-album-cloud-microservices/common/storage/config"
)

type TencentCOS struct {
	client     *cos.Client
	bucket     string
	dispatcher events.Dispatcher
}

// NewTencentCOS 创建tencent OSS 实例
func NewTencentCOS(config *config.StorageConfig, dispatcher events.Dispatcher) (*TencentCOS, error) {
	return nil, nil
}
