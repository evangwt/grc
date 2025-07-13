// Package implementations contains reference cache implementations
// To use this implementation, add to your go.mod:
// go get github.com/bradfitz/gomemcache

// +build ignore

package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
	
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/evangwt/grc"
)

// MemcachedCache demonstrates how to implement grc.CacheClient for Memcached
type MemcachedCache struct {
	client *memcache.Client
}

// NewMemcachedCache creates a new Memcached cache client
func NewMemcachedCache(servers ...string) *MemcachedCache {
	return &MemcachedCache{
		client: memcache.New(servers...),
	}
}

// Get retrieves a value from Memcached
func (m *MemcachedCache) Get(ctx context.Context, key string) (interface{}, error) {
	item, err := m.client.Get(key)
	if err != nil {
		if err == memcache.ErrCacheMiss {
			return nil, grc.ErrCacheMiss
		}
		return nil, err
	}
	return item.Value, nil
}

// Set stores a value in Memcached with TTL
func (m *MemcachedCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	item := &memcache.Item{
		Key:        key,
		Value:      data,
		Expiration: int32(ttl.Seconds()),
	}
	
	return m.client.Set(item)
}

// Close closes the Memcached connection (if needed)
func (m *MemcachedCache) Close() error {
	// Memcached client doesn't need explicit closing
	return nil
}