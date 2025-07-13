package grc

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// testMemoryCache is a simple test implementation for internal testing only
// Users should implement their own cache backends using the examples
type testMemoryCache struct {
	data map[string]*testCacheItem
	mu   sync.RWMutex
}

type testCacheItem struct {
	value  []byte
	expiry time.Time
}

// newTestMemoryCache creates a test cache for testing purposes only
func newTestMemoryCache() *testMemoryCache {
	return &testMemoryCache{
		data: make(map[string]*testCacheItem),
	}
}

// Get retrieves a value from the test cache
func (m *testMemoryCache) Get(ctx context.Context, key string) (interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	item, exists := m.data[key]
	if !exists {
		return nil, ErrCacheMiss
	}

	// Check if expired
	if time.Now().After(item.expiry) {
		return nil, ErrCacheMiss
	}

	return item.value, nil
}

// Set stores a value in the test cache with TTL
func (m *testMemoryCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.data[key] = &testCacheItem{
		value:  data,
		expiry: time.Now().Add(ttl),
	}

	return nil
}