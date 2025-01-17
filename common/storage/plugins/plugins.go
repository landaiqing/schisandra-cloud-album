package plugins

import (
	"schisandra-album-cloud-microservices/common/storage/config"
	"schisandra-album-cloud-microservices/common/storage/constants"
	"schisandra-album-cloud-microservices/common/storage/events"
	"schisandra-album-cloud-microservices/common/storage/manager"
	"schisandra-album-cloud-microservices/common/storage/storage"
)

// pluginFactories 存储所有插件的工厂函数
var pluginFactories = map[string]manager.Factory{
	constants.ProviderAliOSS: func(config *config.StorageConfig, dispatcher events.Dispatcher) (storage.Service, error) {
		return storage.NewAliOSS(config, dispatcher)
	},
}

// RegisterPlugins 注册所有插件
func RegisterPlugins(manager *manager.Manager) error {
	for provider, factory := range pluginFactories {
		if err := manager.RegisterStorage(provider, factory); err != nil {
			return err
		}
	}
	return nil
}
