# Cache Implementation Examples

This directory contains **reference implementations** demonstrating how to create custom cache backends for grc.

⚠️ **Important**: These are reference implementations for learning and development. For production use, we recommend integrating with your preferred cache libraries (go-redis, go-cache, BigCache, etc.).

## Abstract Interface

grc provides a simple and clean `CacheClient` interface that you can implement:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

## Why Use Your Own Implementations?

1. **Leverage existing libraries**: Use battle-tested cache libraries like go-redis, go-cache, BigCache
2. **Match your infrastructure**: Integrate with your existing cache infrastructure  
3. **Get better features**: Established libraries often have more features, optimizations, and community support
4. **Production readiness**: Popular libraries are thoroughly tested in production environments

## Reference Implementations

### 1. MemoryCache (`memory_cache.go`)

A production-ready in-memory cache implementation with automatic cleanup. Good for:
- Development and testing
- Single-instance applications  
- Learning how to implement the interface

**Usage:**
```go
package main

import (
    "time"
    "github.com/evangwt/grc"
    "github.com/evangwt/grc/examples/implementations"
)

func main() {
    // Create memory cache instance (reference implementation)
    memCache := implementations.NewMemoryCache()
    defer memCache.Close()
    
    // Create grc cache with memory backend
    cache := grc.NewGormCache("memory_cache", memCache, grc.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })
    
    // Use with gorm...
}
```

### 2. SimpleRedisClient (`simple_redis_client.go`)

A Redis client implementation without external dependencies. Good for:
- Learning Redis protocol
- Simple Redis usage without external libraries
- Development environments

**Usage:**
```go
redisClient, err := implementations.NewSimpleRedisClient(implementations.SimpleRedisConfig{
    Addr: "localhost:6379",
})
defer redisClient.Close()

cache := grc.NewGormCache("redis_cache", redisClient, grc.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

### 3. MemcachedCache (`memcached_cache.go`)

A Memcached cache implementation using `github.com/bradfitz/gomemcache`:

**Usage:**
```go
memcachedCache := implementations.NewMemcachedCache("localhost:11211")
cache := grc.NewGormCache("memcached_cache", memcachedCache, grc.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

### 4. FileCache (`file_cache.go`)

A file-based cache implementation for persistent local caching:

**Usage:**
```go
fileCache, err := implementations.NewFileCache("/tmp/grc_cache")
if err != nil {
    log.Fatal(err)
}
cache := grc.NewGormCache("file_cache", fileCache, grc.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

## Popular Library Integration Examples

### go-redis Integration

```go
import "github.com/go-redis/redis/v8"

type GoRedisCache struct {
    client *redis.Client
}

func NewGoRedisCache(addr string) *GoRedisCache {
    rdb := redis.NewClient(&redis.Options{
        Addr: addr,
    })
    return &GoRedisCache{client: rdb}
}

func (r *GoRedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, grc.ErrCacheMiss
        }
        return nil, err
    }
    return []byte(val), nil
}

func (r *GoRedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, data, ttl).Err()
}
```

### go-cache Integration

```go
import "github.com/patrickmn/go-cache"

type GoCacheWrapper struct {
    cache *cache.Cache
}

func NewGoCacheWrapper(defaultExpiration, cleanupInterval time.Duration) *GoCacheWrapper {
    c := cache.New(defaultExpiration, cleanupInterval)
    return &GoCacheWrapper{cache: c}
}

func (g *GoCacheWrapper) Get(ctx context.Context, key string) (interface{}, error) {
    if value, found := g.cache.Get(key); found {
        return value, nil
    }
    return nil, grc.ErrCacheMiss
}

func (g *GoCacheWrapper) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    g.cache.Set(key, data, ttl)
    return nil
}
```

## Creating Custom Implementations

You can easily create your own cache backends by implementing the `grc.CacheClient` interface.

### Key Requirements

1. **Error Handling**: Always return `grc.ErrCacheMiss` for cache misses to ensure consistent behavior
2. **Serialization**: Use `json.Marshal/Unmarshal` for data serialization to maintain compatibility
3. **Context Support**: Respect the context parameter for cancellation and timeout handling
4. **Thread Safety**: Ensure your implementation is thread-safe for concurrent access
5. **TTL Handling**: Properly implement TTL behavior according to your storage backend's capabilities

### Example: Custom Database Cache

```go
package main

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"
    "github.com/evangwt/grc"
)

type DatabaseCache struct {
    db *sql.DB
}

func NewDatabaseCache(db *sql.DB) *DatabaseCache {
    // Create cache table if not exists
    db.Exec(`CREATE TABLE IF NOT EXISTS cache_entries (
        cache_key TEXT PRIMARY KEY,
        cache_value BLOB,
        expires_at TIMESTAMP
    )`)
    
    return &DatabaseCache{db: db}
}

func (d *DatabaseCache) Get(ctx context.Context, key string) (interface{}, error) {
    var value []byte
    var expiresAt time.Time
    
    err := d.db.QueryRowContext(ctx, 
        "SELECT cache_value, expires_at FROM cache_entries WHERE cache_key = ?", 
        key).Scan(&value, &expiresAt)
    
    if err != nil {
        if err == sql.ErrNoRows {
            return nil, grc.ErrCacheMiss
        }
        return nil, err
    }
    
    // Check if expired
    if time.Now().After(expiresAt) {
        // Clean up expired entry
        d.db.ExecContext(ctx, "DELETE FROM cache_entries WHERE cache_key = ?", key)
        return nil, grc.ErrCacheMiss
    }
    
    return value, nil
}

func (d *DatabaseCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    
    expiresAt := time.Now().Add(ttl)
    
    _, err = d.db.ExecContext(ctx,
        "INSERT OR REPLACE INTO cache_entries (cache_key, cache_value, expires_at) VALUES (?, ?, ?)",
        key, data, expiresAt)
    
    return err
}
```

## Best Practices

1. **Use established libraries**: Prefer go-redis, go-cache, BigCache over reference implementations
2. **Import the interface only**: Only import `github.com/evangwt/grc` for the interface  
3. **Handle context cancellation**: Check `ctx.Done()` in long-running operations
4. **Implement cleanup**: Provide mechanisms to clean up expired entries
5. **Error wrapping**: Wrap errors with context for better debugging
6. **Resource management**: Implement `Close()` method if your cache needs cleanup
7. **Production testing**: Test your implementation thoroughly in production-like environments

## Testing Your Implementation

Test your cache implementation against the abstract interface:

```go
func TestYourCacheImplementation(t *testing.T) {
    var client grc.CacheClient = NewYourCache()
    
    ctx := context.Background()
    
    // Test cache miss
    _, err := client.Get(ctx, "nonexistent")
    assert.Equal(t, grc.ErrCacheMiss, err)
    
    // Test set/get
    err = client.Set(ctx, "key", "value", time.Minute)
    assert.NoError(t, err)
    
    result, err := client.Get(ctx, "key")
    assert.NoError(t, err)
    assert.NotNil(t, result)
}
```

## Why This Approach?

The grc library follows the **interface segregation principle** - it provides a clean interface and lets you choose the implementation that best fits your needs. This approach:

- **Reduces vendor lock-in**: Switch cache implementations without changing application code
- **Enables testing**: Easy to mock and test with different backends  
- **Promotes reusability**: Same caching logic works with any storage backend
- **Encourages best practices**: Use production-ready libraries instead of reinventing the wheel

**Remember**: The reference implementations here are for learning and development. For production, integrate with your preferred cache libraries!