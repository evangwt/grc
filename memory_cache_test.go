package grc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewMemoryCache(t *testing.T) {
	cache := NewMemoryCache()
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	// Test cache miss
	_, err := cache.Get(ctx, "missing")
	assert.Equal(t, ErrCacheMiss, err)

	// Test set and get
	err = cache.Set(ctx, "key1", "value1", time.Minute)
	assert.NoError(t, err)

	value, err := cache.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"value1\""), value) // JSON marshaled

	// Test TTL expiration
	err = cache.Set(ctx, "expiring", "value", time.Millisecond)
	assert.NoError(t, err)

	time.Sleep(2 * time.Millisecond)
	_, err = cache.Get(ctx, "expiring")
	assert.Equal(t, ErrCacheMiss, err)

	// Test size
	cache.Set(ctx, "size1", "value", time.Minute)
	cache.Set(ctx, "size2", "value", time.Minute)
	// expiring key might still be there until cleanup, so check that size is at least 3
	assert.GreaterOrEqual(t, cache.Size(), 3) // key1 + size1 + size2 (expiring may still be there)
}

func TestMemoryCacheClose(t *testing.T) {
	cache := NewMemoryCache()
	require.NotNil(t, cache)

	ctx := context.Background()

	// Set a value
	err := cache.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)

	// Close the cache
	err = cache.Close()
	assert.NoError(t, err)

	// Operations after close should fail gracefully
	err = cache.Set(ctx, "key2", "value2", time.Minute)
	assert.Equal(t, ErrCacheMiss, err)
}

func TestMemoryCacheCleanup(t *testing.T) {
	cache := NewMemoryCache()
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	// Set values with very short TTL
	cache.Set(ctx, "short1", "value", time.Millisecond)
	cache.Set(ctx, "short2", "value", time.Millisecond)
	cache.Set(ctx, "long", "value", time.Hour)

	// Wait for expiration
	time.Sleep(2 * time.Millisecond)

	// Trigger cleanup by calling cleanupExpired directly
	cache.cleanupExpired()

	// Only the long-lived item should remain
	assert.Equal(t, 1, cache.Size())
}