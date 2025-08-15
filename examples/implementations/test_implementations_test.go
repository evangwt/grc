package implementations

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/evangwt/grc"
)

// Test the implementations in this package
func TestMemoryCacheImplementation(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()
	
	ctx := context.Background()
	
	// Test cache miss
	_, err := cache.Get(ctx, "nonexistent")
	assert.Equal(t, grc.ErrCacheMiss, err)
	
	// Test set/get
	err = cache.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)
	
	result, err := cache.Get(ctx, "key")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestSimpleRedisClientImplementation(t *testing.T) {
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
	
	// Test cache miss
	_, err = client.Get(ctx, "nonexistent")
	assert.Equal(t, grc.ErrCacheMiss, err)
	
	// Test set/get
	err = client.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)
	
	result, err := client.Get(ctx, "key")
	assert.NoError(t, err)
	assert.NotNil(t, result)
}