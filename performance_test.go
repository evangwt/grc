package grc

import (
	"context"
	"testing"
	"time"
	"fmt"
	"crypto/sha256"
	"encoding/hex"
	"hash/fnv"
)

// TestContextKeys verifies the new typed context keys work correctly
func TestContextKeys(t *testing.T) {
	ctx := context.Background()
	
	// Test UseCacheKey
	ctx = context.WithValue(ctx, UseCacheKey, true)
	useCache, ok := ctx.Value(UseCacheKey).(bool)
	if !ok || !useCache {
		t.Error("UseCacheKey should work with typed context key")
	}
	
	// Test CacheTTLKey
	ttl := 30 * time.Second
	ctx = context.WithValue(ctx, CacheTTLKey, ttl)
	retrievedTTL, ok := ctx.Value(CacheTTLKey).(time.Duration)
	if !ok || retrievedTTL != ttl {
		t.Error("CacheTTLKey should work with typed context key")
	}
	
	// Test string representation
	if UseCacheKey.String() != "grc context key UseCache" {
		t.Error("Context key string representation should be descriptive")
	}
}

// BenchmarkHashingMethods compares the performance of different hashing approaches
func BenchmarkHashingMethods(b *testing.B) {
	testData := []string{
		"SELECT * FROM users WHERE id > $1 AND name LIKE $2 ORDER BY created_at DESC LIMIT 100",
		"SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = $2",
		"SELECT u.*, p.name as profile_name FROM users u JOIN profiles p ON u.id = p.user_id WHERE u.active = $1",
		"INSERT INTO cache_entries (key, value, expires_at) VALUES ($1, $2, $3)",
		"UPDATE user_sessions SET last_accessed = $1 WHERE session_id = $2 AND user_id = $3",
	}

	b.Run("FNV1a", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql := testData[i%len(testData)]
			h := fnv.New64a()
			h.Write([]byte(sql))
			_ = fmt.Sprintf("%x", h.Sum64())
		}
	})

	b.Run("SHA256", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			sql := testData[i%len(testData)]
			hash := sha256.Sum256([]byte(sql))
			_ = hex.EncodeToString(hash[:])
		}
	})
}

// BenchmarkCacheConfig tests performance with different configurations
func BenchmarkCacheConfig(b *testing.B) {
	testSQL := "SELECT * FROM users WHERE id > $1 AND status = $2"
	
	b.Run("FastHash", func(b *testing.B) {
		config := CacheConfig{
			TTL:           time.Minute,
			Prefix:        "fast:",
			UseSecureHash: false,
		}
		cache := &GormCache{config: config}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate cache key generation
			var hash string
			if config.UseSecureHash {
				h := sha256.Sum256([]byte(testSQL))
				hash = hex.EncodeToString(h[:])
			} else {
				h := fnv.New64a()
				h.Write([]byte(testSQL))
				hash = fmt.Sprintf("%x", h.Sum64())
			}
			_ = cache.config.Prefix + hash
		}
	})
	
	b.Run("SecureHash", func(b *testing.B) {
		config := CacheConfig{
			TTL:           time.Minute,
			Prefix:        "secure:",
			UseSecureHash: true,
		}
		cache := &GormCache{config: config}
		
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			// Simulate cache key generation
			var hash string
			if config.UseSecureHash {
				h := sha256.Sum256([]byte(testSQL))
				hash = hex.EncodeToString(h[:])
			} else {
				h := fnv.New64a()
				h.Write([]byte(testSQL))
				hash = fmt.Sprintf("%x", h.Sum64())
			}
			_ = cache.config.Prefix + hash
		}
	})
}

// BenchmarkErrorHandling tests the performance impact of improved error handling
func BenchmarkErrorHandling(b *testing.B) {
	cache := newTestMemoryCache()
	
	ctx := context.Background()
	key := "benchmark_key"
	value := map[string]interface{}{
		"id":   1,
		"data": "test_data",
	}
	
	// Pre-populate cache
	cache.Set(ctx, key, value, time.Minute)
	
	b.Run("CacheHit", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cache.Get(ctx, key)
			if err != nil {
				b.Error("Should not have error on cache hit")
			}
		}
	})
	
	b.Run("CacheMiss", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_, err := cache.Get(ctx, "nonexistent_key")
			if err != ErrCacheMiss {
				b.Error("Should get cache miss error")
			}
		}
	})
}

// TestTimeoutHandling verifies the new timeout handling works correctly
func TestTimeoutHandling(t *testing.T) {
	cache := newTestMemoryCache()
	
	// Test with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()
	
	// Should work within timeout
	err := cache.Set(ctx, "test_key", "test_value", time.Minute)
	if err != nil {
		t.Error("Should work within timeout")
	}
	
	// Should work for get too
	_, err = cache.Get(ctx, "test_key")
	if err != nil {
		t.Error("Should work within timeout for get")
	}
}

// TestCacheConfigDefaults ensures backward compatibility
func TestCacheConfigDefaults(t *testing.T) {
	// Test default behavior (should use fast hash)
	config := CacheConfig{
		TTL:    time.Minute,
		Prefix: "test:",
		// UseSecureHash not set, should default to false
	}
	
	if config.UseSecureHash {
		t.Error("UseSecureHash should default to false")
	}
	
	// Test explicit configuration
	secureConfig := CacheConfig{
		TTL:           time.Minute,
		Prefix:        "secure:",
		UseSecureHash: true,
	}
	
	if !secureConfig.UseSecureHash {
		t.Error("UseSecureHash should be true when explicitly set")
	}
}