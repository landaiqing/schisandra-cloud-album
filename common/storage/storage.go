package storage

import (
	"schisandra-album-cloud-microservices/common/storage/events"
	"schisandra-album-cloud-microservices/common/storage/manager"
	"schisandra-album-cloud-microservices/common/storage/plugins"
)

// InitStorageManager 初始化存储管理器
func InitStorageManager() *manager.Manager {
	// 初始化事件分发器
	dispatcher := events.NewDispatcher()

	// 初始化存储管理器
	m := manager.NewStorageManager(dispatcher)

	// 注册插件
	if err := plugins.RegisterPlugins(m); err != nil {
		panic(err)
		return nil
	}
	return m
}
