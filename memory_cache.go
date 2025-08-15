package grc

import (
	"context"
	"encoding/json"
	"sync"
	"time"
)

// MemoryCache is a production-ready in-memory cache implementation
// It provides thread-safe operations and automatic cleanup of expired items
type MemoryCache struct {
	data     map[string]*memoryCacheItem
	mu       sync.RWMutex
	stopChan chan struct{}
	stopped  bool
}

type memoryCacheItem struct {
	value  []byte
	expiry time.Time
}

// NewMemoryCache creates a new in-memory cache instance with automatic cleanup
func NewMemoryCache() *MemoryCache {
	mc := &MemoryCache{
		data:     make(map[string]*memoryCacheItem),
		stopChan: make(chan struct{}),
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
		return nil, ErrCacheMiss
	}

	// Check if expired
	if time.Now().After(item.expiry) {
		return nil, ErrCacheMiss
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

	// Check if cache is stopped
	if m.stopped {
		return ErrCacheMiss // Return cache miss to indicate cache is not operational
	}

	m.data[key] = &memoryCacheItem{
		value:  data,
		expiry: time.Now().Add(ttl),
	}

	return nil
}

// Close stops the cleanup goroutine and clears the cache
func (m *MemoryCache) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.stopped {
		m.stopped = true
		close(m.stopChan)
		m.data = nil
	}
	return nil
}

// cleanup removes expired items from the cache periodically
func (m *MemoryCache) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.cleanupExpired()
		case <-m.stopChan:
			return
		}
	}
}

// cleanupExpired removes expired items (internal method)
func (m *MemoryCache) cleanupExpired() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.stopped {
		return
	}

	now := time.Now()
	for key, item := range m.data {
		if now.After(item.expiry) {
			delete(m.data, key)
		}
	}
}

// Size returns the current number of items in the cache
func (m *MemoryCache) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.data)
}