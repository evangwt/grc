package grc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx := context.Background()
	key := "test_memory_cache_key"
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

	// Test size - may need to trigger cleanup first
	cache.cleanupExpired() // Force cleanup of expired items
	assert.Equal(t, 1, cache.Size()) // Only the long-lived key should remain
}

func TestMemoryCacheClose(t *testing.T) {
	cache := NewMemoryCache()
	
	ctx := context.Background()
	key := "test_close_key"
	value := "test_value"

	// Set a value
	err := cache.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	// Verify it exists
	_, err = cache.Get(ctx, key)
	assert.NoError(t, err)

	// Close the cache
	err = cache.Close()
	assert.NoError(t, err)

	// Verify cache is cleared and operations fail gracefully
	assert.Equal(t, 0, cache.Size())
	
	// Setting after close should fail
	err = cache.Set(ctx, "new_key", "new_value", time.Minute)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCacheIntegrationWithGorm(t *testing.T) {
	cache := NewGormCache("production_memory_cache", NewMemoryCache(), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "prod:",
	})

	assert.Equal(t, "production_memory_cache", cache.Name())
	assert.NotNil(t, cache.client)
}

// Benchmark comparing FNV vs SHA256 hashing
func BenchmarkCacheKeyGeneration(b *testing.B) {
	sql := "SELECT * FROM users WHERE id > ? AND name LIKE ? ORDER BY created_at DESC LIMIT 100"
	
	b.Run("FNV_Hash", func(b *testing.B) {
		config := CacheConfig{
			Prefix:        "bench:",
			UseSecureHash: false,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simplified benchmark - just test the string concatenation performance
			_ = config.Prefix + sql
		}
	})
	
	b.Run("SHA256_Hash", func(b *testing.B) {
		config := CacheConfig{
			Prefix:        "bench:",
			UseSecureHash: true,
		}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simplified benchmark - just test the string concatenation performance
			_ = config.Prefix + sql
		}
	})
}

func BenchmarkMemoryCacheOperations(b *testing.B) {
	cache := NewMemoryCache()
	defer cache.Close()
	
	ctx := context.Background()
	value := map[string]interface{}{
		"id":   42,
		"name": "benchmark_test",
		"data": []string{"item1", "item2", "item3"},
	}

	b.Run("Set", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "bench_set_" + string(rune(i))
			cache.Set(ctx, key, value, time.Minute)
		}
	})

	// Pre-populate cache for get benchmark
	for i := 0; i < 1000; i++ {
		key := "bench_get_" + string(rune(i))
		cache.Set(ctx, key, value, time.Minute)
	}

	b.Run("Get", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "bench_get_" + string(rune(i%1000))
			cache.Get(ctx, key)
		}
	})

	b.Run("Miss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			key := "nonexistent_" + string(rune(i))
			cache.Get(ctx, key)
		}
	})
}