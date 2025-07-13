package grc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMemoryCache(t *testing.T) {
	cache := newTestMemoryCache()

	ctx := context.Background()
	key := "test_key"
	value := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	// Test cache miss
	_, err := cache.Get(ctx, key)
	assert.Equal(t, ErrCacheMiss, err)

	// Test set and get
	err = cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	result, err := cache.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test expiration
	shortKey := "short_key"
	err = cache.Set(ctx, shortKey, value, time.Millisecond*10)
	assert.NoError(t, err)

	// Should get immediately
	result, err = cache.Get(ctx, shortKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Wait for expiration
	time.Sleep(time.Millisecond * 20)
	_, err = cache.Get(ctx, shortKey)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCacheIntegration(t *testing.T) {
	// This test demonstrates how to use custom cache implementations
	cache := NewGormCache("test_cache", newTestMemoryCache(), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "test:",
	})

	assert.Equal(t, "test_cache", cache.Name())
	assert.NotNil(t, cache.client)
}