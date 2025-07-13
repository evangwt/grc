package implementations

import (
	"context"
	"encoding/json"
	"sync"
	"time"
	
	"github.com/evangwt/grc"
)

// MemoryCache is a reference implementation of an in-memory cache
// This demonstrates how to implement the grc.CacheClient interface
type MemoryCache struct {
	data map[string]*cacheItem
	mu   sync.RWMutex
}

type cacheItem struct {
	value  []byte
	expiry time.Time
}

// NewMemoryCache creates a new in-memory cache instance
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		data: make(map[string]*cacheItem),
	}
	// Start cleanup goroutine
	go mc.cleanup()
	return mc
}

// Get retrieves a value from the memory cache
func (m *MemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, grc.ErrCacheMiss
	}

	// Check if expired
	if time.Now().After(item.expiry) {
		return nil, grc.ErrCacheMiss
	}

	return item.value, nil
}

// Set stores a value in the memory cache with TTL
func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = &cacheItem{
		value:  data,
		expiry: time.Now().Add(ttl),
	}

	return nil
}

// cleanup removes expired items from the cache
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.mu.Lock()
			now := time.Now()
			for key, item := range m.data {
				if now.After(item.expiry) {
					delete(m.data, key)
				}
			}
			m.mu.Unlock()
		}
	}
}