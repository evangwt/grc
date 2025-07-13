package grc

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

// TestAbstractCacheInterface demonstrates the abstract nature of the CacheClient interface
// This test shows that any implementation of CacheClient can be used interchangeably
func TestAbstractCacheInterface(t *testing.T) {
	// Test function that works with any CacheClient implementation
	testCacheBehavior := func(t *testing.T, client CacheClient, description string) {
		ctx := context.Background()
		
		// Test data
		testCases := []struct {
			key   string
			value interface{}
			ttl   time.Duration
		}{
			{"string_key", "simple string value", time.Minute},
			{"number_key", 42, time.Minute},
			{"object_key", map[string]interface{}{
				"id":   123,
				"name": "test object",
				"tags": []string{"cache", "test"},
			}, time.Minute},
			{"array_key", []interface{}{"item1", "item2", "item3"}, time.Minute},
		}

		for _, tc := range testCases {
			t.Run(description+"_"+tc.key, func(t *testing.T) {
				// Test cache miss first
				_, err := client.Get(ctx, tc.key)
				assert.Equal(t, ErrCacheMiss, err, "Should get cache miss for non-existent key")

				// Set value
				err = client.Set(ctx, tc.key, tc.value, tc.ttl)
				assert.NoError(t, err, "Should be able to set value")

				// Get value back
				result, err := client.Get(ctx, tc.key)
				assert.NoError(t, err, "Should be able to get value")
				assert.NotNil(t, result, "Result should not be nil")
			})
		}
	}

	// Test with test memory cache
	t.Run("MemoryCache", func(t *testing.T) {
		memoryCache := newTestMemoryCache()
		testCacheBehavior(t, memoryCache, "memory")
	})

	// Test with SimpleRedisClient using miniredis
	t.Run("SimpleRedisClient", func(t *testing.T) {
		server := miniredis.RunT(t)
		defer server.Close()

		config := SimpleRedisConfig{
			Addr:     server.Addr(),
			Password: "",
			DB:       0,
		}

		redisClient, err := NewSimpleRedisClient(config)
		assert.NoError(t, err)
		defer redisClient.Close()

		testCacheBehavior(t, redisClient, "redis")
	})
}

// TestGormCacheWithDifferentBackends shows how GormCache works with different storage backends
func TestGormCacheWithDifferentBackends(t *testing.T) {
	testGormCacheSetup := func(t *testing.T, client CacheClient, name string) {
		config := CacheConfig{
			TTL:    30 * time.Second,
			Prefix: "test:",
		}

		cache := NewGormCache("test_cache_"+name, client, config)
		
		assert.Equal(t, "test_cache_"+name, cache.Name())
		assert.NotNil(t, cache.client)
		
		// Test that the cache implements the expected interface
		assert.Implements(t, (*interface{ Name() string })(nil), cache)
	}

	// Test GormCache with test memory cache backend
	t.Run("WithMemoryCache", func(t *testing.T) {
		memoryCache := newTestMemoryCache()
		testGormCacheSetup(t, memoryCache, "memory")
	})

	// Test GormCache with SimpleRedisClient backend
	t.Run("WithSimpleRedisClient", func(t *testing.T) {
		server := miniredis.RunT(t)
		defer server.Close()

		config := SimpleRedisConfig{
			Addr:     server.Addr(),
			Password: "",
			DB:       0,
		}

		redisClient, err := NewSimpleRedisClient(config)
		assert.NoError(t, err)
		defer redisClient.Close()

		testGormCacheSetup(t, redisClient, "redis")
	})
}

// TestCacheClientErrorHandling ensures both implementations handle errors consistently
func TestCacheClientErrorHandling(t *testing.T) {
	testErrorHandling := func(t *testing.T, client CacheClient, description string) {
		ctx := context.Background()
		
		// Test cache miss behavior
		_, err := client.Get(ctx, "nonexistent_key_"+description)
		assert.Equal(t, ErrCacheMiss, err, "Should return ErrCacheMiss for non-existent keys")
		
		// Test setting and getting with positive TTL
		err = client.Set(ctx, "positive_ttl_key", "test_value", time.Second)
		assert.NoError(t, err, "Should handle positive TTL gracefully")
		
		result, err := client.Get(ctx, "positive_ttl_key")
		assert.NoError(t, err, "Should be able to get value with positive TTL")
		assert.NotNil(t, result, "Result should not be nil")
	}

	// Test test memory cache error handling
	t.Run("MemoryCache", func(t *testing.T) {
		memoryCache := newTestMemoryCache()
		testErrorHandling(t, memoryCache, "memory")
	})

	// Test SimpleRedisClient error handling
	t.Run("SimpleRedisClient", func(t *testing.T) {
		server := miniredis.RunT(t)
		defer server.Close()

		config := SimpleRedisConfig{
			Addr:     server.Addr(),
			Password: "",
			DB:       0,
		}

		redisClient, err := NewSimpleRedisClient(config)
		assert.NoError(t, err)
		defer redisClient.Close()

		testErrorHandling(t, redisClient, "redis")
	})
}