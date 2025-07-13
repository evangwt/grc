# Cache Implementation Examples

This directory contains reference implementations demonstrating how to create custom cache backends for grc.

## Abstract Interface

grc provides a simple and clean `CacheClient` interface that you can implement:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

## Reference Implementations

### 1. MemoryCache (`memory_cache.go`)

A thread-safe in-memory cache implementation perfect for:
- Development and testing
- Single-instance applications
- Temporary caching needs

**Usage:**
```go
package main

import (
    "time"
    "github.com/evangwt/grc"
    "github.com/evangwt/grc/examples/implementations"
)

func main() {
    // Create memory cache instance
    memCache := implementations.NewMemoryCache()
    
    // Create grc cache with memory backend
    cache := grc.NewGormCache("memory_cache", memCache, grc.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })
    
    // Use with gorm...
}
```

### 2. MemcachedCache (`memcached_cache.go`)

A Memcached cache implementation using `github.com/bradfitz/gomemcache`:

**Usage:**
```go
memcachedCache := implementations.NewMemcachedCache("localhost:11211")
cache := grc.NewGormCache("memcached_cache", memcachedCache, grc.CacheConfig{
    TTL:    60 * time.Second,
    Prefix: "cache:",
})
```

### 3. FileCache (`file_cache.go`)

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

1. **Import the interface only**: Only import `github.com/evangwt/grc` for the interface
2. **Handle context cancellation**: Check `ctx.Done()` in long-running operations
3. **Implement cleanup**: Provide mechanisms to clean up expired entries
4. **Error wrapping**: Wrap errors with context for better debugging
5. **Resource management**: Implement `Close()` method if your cache needs cleanup

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