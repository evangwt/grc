package grc

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRedisClient(t *testing.T) {
	// Use miniredis for testing
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	config := RedisConfig{
		Addr: server.Addr(),
	}

	client, err := NewRedisClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()

	// Test cache miss
	_, err = client.Get(ctx, "missing")
	assert.Equal(t, ErrCacheMiss, err)

	// Test set and get
	err = client.Set(ctx, "key1", "value1", time.Minute)
	assert.NoError(t, err)

	value, err := client.Get(ctx, "key1")
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"value1\""), value) // JSON marshaled
}

func TestRedisClientWithAuth(t *testing.T) {
	// Use miniredis for testing
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	// Set password on miniredis
	server.RequireAuth("testpass")

	config := RedisConfig{
		Addr:     server.Addr(),
		Password: "testpass",
	}

	client, err := NewRedisClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()

	// Test operations work with auth
	err = client.Set(ctx, "auth_key", "auth_value", time.Minute)
	assert.NoError(t, err)

	value, err := client.Get(ctx, "auth_key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"auth_value\""), value)
}

func TestRedisClientWithDatabase(t *testing.T) {
	// Use miniredis for testing
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	config := RedisConfig{
		Addr: server.Addr(),
		DB:   1, // Use database 1
	}

	client, err := NewRedisClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()

	// Test operations work with specific database
	err = client.Set(ctx, "db_key", "db_value", time.Minute)
	assert.NoError(t, err)

	value, err := client.Get(ctx, "db_key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"db_value\""), value)
}

func TestRedisClientTTL(t *testing.T) {
	// Use miniredis for testing
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	config := RedisConfig{
		Addr: server.Addr(),
	}

	client, err := NewRedisClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)
	defer client.Close()

	ctx := context.Background()

	// Test TTL functionality
	err = client.Set(ctx, "ttl_key", "ttl_value", time.Second)
	assert.NoError(t, err)

	// Should exist immediately
	value, err := client.Get(ctx, "ttl_key")
	assert.NoError(t, err)
	assert.Equal(t, []byte("\"ttl_value\""), value)

	// Fast forward time in miniredis
	server.FastForward(2 * time.Second)

	// Should be expired
	_, err = client.Get(ctx, "ttl_key")
	assert.Equal(t, ErrCacheMiss, err)
}

func TestRedisClientClose(t *testing.T) {
	// Use miniredis for testing
	server, err := miniredis.Run()
	require.NoError(t, err)
	defer server.Close()

	config := RedisConfig{
		Addr: server.Addr(),
	}

	client, err := NewRedisClient(config)
	require.NoError(t, err)
	require.NotNil(t, client)

	ctx := context.Background()

	// Set a value
	err = client.Set(ctx, "key", "value", time.Minute)
	assert.NoError(t, err)

	// Close the client
	err = client.Close()
	assert.NoError(t, err)

	// Note: Since Redis client reconnects automatically, we just verify Close() works
	// The connection will be re-established on next operation in this simple implementation
}