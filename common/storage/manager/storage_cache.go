package manager

import (
	"schisandra-album-cloud-microservices/common/storage/storage"
	"sync"
	"time"
)

// CacheEntry 缓存项定义
type CacheEntry struct {
	Instance storage.Service
	mu       sync.Mutex // 确保 LastUsed 的线程安全
	LastUsed time.Time
}

// UserStorageCache 管理每个用户的存储实例缓存
type UserStorageCache struct {
	cache sync.Map // map[userID::providerName]*CacheEntry
}

// NewUserStorageCache 创建新的用户存储缓存
func NewUserStorageCache() *UserStorageCache {
	return &UserStorageCache{}
}

// GetOrCreate 获取或创建缓存实例
func (usc *UserStorageCache) GetOrCreate(key, providerName string, factory func() (storage.Service, error)) (storage.Service, error) {
	cacheKey := key + "::" + providerName

	if entry, exists := usc.cache.Load(cacheKey); exists {
		usc.updateLastUsed(entry.(*CacheEntry))
		return entry.(*CacheEntry).Instance, nil
	}

	instance, err := factory()
	if err != nil {
		return nil, err
	}

	cacheEntry := &CacheEntry{
		Instance: instance,
		LastUsed: time.Now(),
	}
	usc.cache.Store(cacheKey, cacheEntry)
	return instance, nil
}

// ClearUnused 清理长时间未使用的实例
func (usc *UserStorageCache) ClearUnused(timeout time.Duration) {
	now := time.Now()
	usc.cache.Range(func(key, value interface{}) bool {
		entry := value.(*CacheEntry)
		entry.mu.Lock()
		defer entry.mu.Unlock()

		if now.Sub(entry.LastUsed) > timeout {
			usc.cache.Delete(key)
		}
		return true
	})
}

// updateLastUsed 更新最后使用时间
func (usc *UserStorageCache) updateLastUsed(entry *CacheEntry) {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	entry.LastUsed = time.Now()
}
