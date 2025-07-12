package grc

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func TestSimpleRedisClient(t *testing.T) {
	// Create a miniredis server for testing
	server := miniredis.RunT(t)
	defer server.Close()

	config := SimpleRedisConfig{
		Addr:     server.Addr(),
		Password: "",
		DB:       0,
	}

	client, err := NewSimpleRedisClient(config)
	assert.NoError(t, err)
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
	err = client.Set(ctx, shortKey, value, time.Second*1)
	assert.NoError(t, err)

	// Should get immediately
	result, err = client.Get(ctx, shortKey)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Fast forward time in miniredis and check expiration
	server.FastForward(time.Second * 2)
	_, err = client.Get(ctx, shortKey)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestSimpleRedisClientIntegration(t *testing.T) {
	// Create a miniredis server for testing
	server := miniredis.RunT(t)
	defer server.Close()

	config := SimpleRedisConfig{
		Addr:     server.Addr(),
		Password: "",
		DB:       0,
	}

	client, err := NewSimpleRedisClient(config)
	assert.NoError(t, err)
	defer client.Close()

	cache := NewGormCache("test_redis_cache", client, CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "test:",
	})

	assert.Equal(t, "test_redis_cache", cache.Name())
	assert.NotNil(t, cache.client)
}

// TestCacheClientInterface demonstrates that both MemoryCache and SimpleRedisClient 
// implement the same CacheClient interface, showing the abstract design
func TestCacheClientInterface(t *testing.T) {
	// Test that both implementations satisfy the CacheClient interface
	var memoryClient CacheClient = NewMemoryCache()
	var redisClient CacheClient

	// Create a miniredis server for testing
	server := miniredis.RunT(t)
	defer server.Close()

	config := SimpleRedisConfig{
		Addr:     server.Addr(),
		Password: "",
		DB:       0,
	}

	client, err := NewSimpleRedisClient(config)
	assert.NoError(t, err)
	defer client.Close()
	redisClient = client

	// Test both implementations with the same interface
	testCacheClient := func(t *testing.T, client CacheClient, name string) {
		ctx := context.Background()
		key := "test_interface_key"
		value := map[string]interface{}{
			"id":   42,
			"name": "interface_test",
		}

		// Test cache miss
		_, err := client.Get(ctx, key)
		assert.Equal(t, ErrCacheMiss, err, "Cache miss test failed for %s", name)

		// Test set and get
		err = client.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err, "Set operation failed for %s", name)

		result, err := client.Get(ctx, key)
		assert.NoError(t, err, "Get operation failed for %s", name)
		assert.NotNil(t, result, "Result should not be nil for %s", name)
	}

	// Test both implementations using the same interface
	t.Run("MemoryCache", func(t *testing.T) {
		testCacheClient(t, memoryClient, "MemoryCache")
	})

	t.Run("SimpleRedisClient", func(t *testing.T) {
		testCacheClient(t, redisClient, "SimpleRedisClient")
	})
}

// TestRedisClientWithPassword tests authentication with miniredis
func TestRedisClientWithPassword(t *testing.T) {
	// Create a miniredis server with password
	server := miniredis.RunT(t)
	defer server.Close()
	
	password := "testpassword"
	server.RequireAuth(password)

	config := SimpleRedisConfig{
		Addr:     server.Addr(),
		Password: password,
		DB:       0,
	}

	client, err := NewSimpleRedisClient(config)
	assert.NoError(t, err)
	defer client.Close()

	ctx := context.Background()
	key := "auth_test_key"
	value := "auth_test_value"

	// Test operations with authentication
	err = client.Set(ctx, key, value, time.Minute)
	assert.NoError(t, err)

	result, err := client.Get(ctx, key)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

// TestRedisClientWithDatabase tests database selection
func TestRedisClientWithDatabase(t *testing.T) {
	// Create a miniredis server
	server := miniredis.RunT(t)
	defer server.Close()

	// Test with different database numbers
	for dbNum := 0; dbNum < 3; dbNum++ {
		config := SimpleRedisConfig{
			Addr:     server.Addr(),
			Password: "",
			DB:       dbNum,
		}

		client, err := NewSimpleRedisClient(config)
		assert.NoError(t, err)

		ctx := context.Background()
		key := "db_test_key"
		value := map[string]interface{}{"db": dbNum}

		err = client.Set(ctx, key, value, time.Minute)
		assert.NoError(t, err)

		result, err := client.Get(ctx, key)
		assert.NoError(t, err)
		assert.NotNil(t, result)

		client.Close()
	}
}