package manager

import (
	"errors"
	"schisandra-album-cloud-microservices/common/storage/config"
	"schisandra-album-cloud-microservices/common/storage/events"
	"schisandra-album-cloud-microservices/common/storage/storage"
	"sync"
	"time"
)

// Factory 定义存储服务工厂函数类型
type Factory func(config *config.StorageConfig, dispatcher events.Dispatcher) (storage.Service, error)

// Manager 管理存储服务的注册、实例化和缓存
type Manager struct {
	mu         sync.RWMutex
	registry   map[string]Factory
	dispatcher events.Dispatcher
	cache      *UserStorageCache
}

// NewStorageManager 创建新的存储管理器
func NewStorageManager(dispatcher events.Dispatcher) *Manager {
	return &Manager{
		registry:   make(map[string]Factory),
		dispatcher: dispatcher,
		cache:      NewUserStorageCache(),
	}
}

// RegisterStorage 注册存储服务提供商
func (sm *Manager) RegisterStorage(provider string, factory Factory) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if provider == "" || factory == nil {
		return errors.New("invalid provider or factory")
	}
	if _, exists := sm.registry[provider]; exists {
		return errors.New("provider already registered")
	}

	sm.registry[provider] = factory
	return nil
}

// GetStorage 获取或创建存储服务实例
func (sm *Manager) GetStorage(key string, config *config.StorageConfig) (storage.Service, error) {
	if key == "" || config.Provider == "" {
		return nil, errors.New("invalid user ID or provider")
	}

	// 尝试从缓存获取实例
	return sm.cache.GetOrCreate(key, config.Provider, func() (storage.Service, error) {
		// 从注册表中查找工厂函数
		sm.mu.RLock()
		factory, exists := sm.registry[config.Provider]
		sm.mu.RUnlock()

		if !exists {
			return nil, errors.New("unsupported provider: " + config.Provider)
		}

		// 创建新实例并返回
		return factory(config, sm.dispatcher)
	})
}

// ClearUnused 清理长时间未使用的缓存实例
func (sm *Manager) ClearUnused(timeout time.Duration) {
	sm.cache.ClearUnused(timeout)
}

// ListProviders 列出所有注册的存储服务提供商
func (sm *Manager) ListProviders() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	providers := make([]string, 0, len(sm.registry))
	for provider := range sm.registry {
		providers = append(providers, provider)
	}
	return providers
}
