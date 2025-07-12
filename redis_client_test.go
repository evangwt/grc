package grc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSimpleRedisClient(t *testing.T) {
	// This test will only run if Redis is available
	// Skip by default to avoid test failures in CI
	t.Skip("Redis integration test - run manually with Redis server available")

	config := SimpleRedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	client, err := NewSimpleRedisClient(config)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test_simple_redis_key"
	value := map[string]interface{}{
		"id":   1,
		"name": "test",
	}

	// Test set and get
	err = client.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	result, err := client.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Test cache miss
	_, err = client.Get(ctx, "nonexistent_key")
	assert.Equal(t, ErrCacheMiss, err)

	// Test expiration
	shortKey := "short_key"
	err = client.Set(ctx, shortKey, value, time.Millisecond*10)
	assert.NoError(t, err)

	// Should get immediately
	result, err = client.Get(ctx, shortKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Wait for expiration
	time.Sleep(time.Millisecond * 20)
	_, err = client.Get(ctx, shortKey)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestSimpleRedisClientIntegration(t *testing.T) {
	// This demonstrates how to use SimpleRedisClient with GormCache
	// This test is mostly for documentation purposes
	config := SimpleRedisConfig{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	}

	// This would fail if Redis is not available, which is expected
	client, err := NewSimpleRedisClient(config)
	if err != nil {
		t.Skipf("Redis not available, skipping integration test: %v", err)
		return
	}
	defer client.Close()

	cache := NewGormCache("test_redis_cache", client, CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "test:",
	})

	assert.Equal(t, "test_redis_cache", cache.Name())
	assert.NotNil(t, cache.client)
}