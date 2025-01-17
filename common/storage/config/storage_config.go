package config

import (
	"errors"
)

// StorageConfig 用户存储配置结构
type StorageConfig struct {
	// 必须字段
	Provider   string `json:"provider"`    // 存储服务提供商
	AccessKey  string `json:"access_key"`  // 访问密钥
	SecretKey  string `json:"secret_key"`  // 安全密钥
	Region     string `json:"region"`      // 区域
	BucketName string `json:"bucket_name"` // 存储桶

	// 可选字段
	Endpoint    string            `json:"endpoint,omitempty"`     // 自定义 API 终端地址
	ExtraConfig map[string]string `json:"extra_config,omitempty"` // 额外的服务商特定配置
}

// Validate 校验存储配置是否有效
func (sc *StorageConfig) Validate() error {
	if sc.Provider == "" {
		return errors.New("provider is required")
	}
	if sc.AccessKey == "" || sc.SecretKey == "" {
		return errors.New("access_key and secret_key are required")
	}
	if sc.Region == "" {
		return errors.New("region is required")
	}
	if sc.BucketName == "" {
		return errors.New("bucket_name is required")
	}
	return nil
}
