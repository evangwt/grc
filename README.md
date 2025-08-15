# grc: a simple gorm cache plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/evangwt/grc)](https://goreportcard.com/report/github.com/evangwt/grc)[![GitHub release](https://img.shields.io/github/release/evangwt/grc.svg)](https://github.com/evangwt/grc/releases/)

grc is a gorm plugin that provides a **简洁优雅的使用方式** (simple and elegant usage) for data caching with a **clean abstract interface** design.

## ✨ Features

- **🎯 Clean Abstract Interface**: Simple `CacheClient` interface for maximum flexibility
- **🔌 Pluggable Architecture**: Implement any cache backend (memory, Redis, Memcached, database, file, etc.)
- **🚀 Zero Required Dependencies**: Core library has no external cache dependencies
- **📝 Simple Context-Based API**: Control cache behavior through gorm session context
- **🧪 Comprehensive Testing**: Full test coverage with production-ready examples
- **⚡ Production Ready**: Thread-safe interface design suitable for high-concurrency
- **📚 Rich Examples**: Reference implementations for common cache backends
- **🏃‍♂️ High Performance**: Optimized hashing (FNV vs SHA256) with 27% performance improvement
- **🔧 Enhanced Error Handling**: Timeout support with graceful error handling
- **🛡️ Type Safety**: Type-safe context keys with proper error definitions

## 🏗️ Architecture

grc implements a **clean abstract interface design** with the `CacheClient` interface:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

This elegant abstraction allows you to:
- **Implement any storage backend** (memory, Redis, Memcached, database, file, etc.)
- **Switch backends** seamlessly without changing your application code
- **Test easily** with different backends for unit/integration tests
- **Extend functionality** with custom cache behaviors
- **Maintain consistency** across different deployment environments

## ⚡ Performance Optimizations

grc includes several performance optimizations for production use:

### Configurable Hashing Strategy
```go
cache := grc.NewGormCache("my_cache", backend, grc.CacheConfig{
    TTL:           60 * time.Second,
    Prefix:        "cache:",
    UseSecureHash: false, // Fast FNV hashing (default)
    // UseSecureHash: true,  // Secure SHA256 hashing (for collision resistance)
})
```

**Benchmark Results:**
- **FNV1a Hashing**: 194.7 ns/op, 24 B/op, 2 allocs/op
- **SHA256 Hashing**: 265.5 ns/op, 217 B/op, 3 allocs/op
- **Performance Gain**: ~27% faster with FNV hashing

### Memory Cache Performance
- **Cache Hit**: ~174 ns/op with minimal allocations
- **Cache Miss**: ~47 ns/op with zero allocations  
- **Cache Set**: ~1536 ns/op for complex objects
- **Automatic Cleanup**: Background cleanup of expired items

### Enhanced Error Handling
- **Timeout Support**: 5-second default timeout for cache operations
- **Graceful Degradation**: Cache failures don't break database queries
- **Type-Safe Errors**: `ErrCacheMiss` and `ErrCacheTimeout` for proper handling

## 📦 Installation

To use grc, you only need gorm installed:

```bash
go get -u gorm.io/gorm
go get -u github.com/evangwt/grc
```

**No external cache dependencies required!** 🎉

## 🚀 Quick Start

### Step 1: Implement Your Cache Backend

The power of grc lies in its interface-first design. You can use any cache implementation that satisfies the `CacheClient` interface:

#### Option 1: Use Popular Cache Libraries

```go
// Example: Wrapping go-redis (popular Redis library)
type GoRedisCache struct {
    client *redis.Client
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

// Example: Wrapping go-cache (popular memory cache)
type GoCacheWrapper struct {
    cache *cache.Cache
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

#### Option 2: Reference Implementations (examples/implementations/)

```go
// Use reference implementations for testing and development
import "github.com/evangwt/grc/examples/implementations"

// Memory cache reference implementation
memCache := implementations.NewMemoryCache()
defer memCache.Close()

// Redis reference implementation (no external dependencies)
redisClient, err := implementations.NewSimpleRedisClient(implementations.SimpleRedisConfig{
    Addr:        "localhost:6379",
    Password:    "", // optional
    DB:          0,  // optional
    MaxIdleTime: 5 * time.Minute, // optional: auto-reconnect after idle time
})
defer redisClient.Close()
```

#### Option 3: Custom Implementation

```go
// Implement your own cache backend
type MyCustomCache struct {
    // your fields
}

func (c *MyCustomCache) Get(ctx context.Context, key string) (interface{}, error) {
    // your implementation
    return nil, grc.ErrCacheMiss // return this for cache misses
}

func (c *MyCustomCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    // your implementation
    return nil
}
```

### Step 2: Setup GormCache

```go
package main

import (
    "context"
    "time"
    "github.com/evangwt/grc"
    "github.com/evangwt/grc/examples/implementations"
    "gorm.io/gorm"
)

func main() {
    // Initialize your chosen cache backend
    // Use YOUR implementation - the examples are just references!
    cacheBackend := implementations.NewMemoryCache() // Reference implementation for development
    
    // Create grc cache with performance optimizations
    cache := grc.NewGormCache("my_cache", cacheBackend, grc.CacheConfig{
        TTL:           60 * time.Second,
        Prefix:        "cache:",
        UseSecureHash: false, // Use fast FNV hashing (27% faster than SHA256)
    })
    
    // Register with gorm
    db.Use(cache)
}
```

### Step 3: Use with Gorm

```go
// Enable cache for a query
ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
db.Session(&gorm.Session{Context: ctx}).Find(&users)

// Use custom TTL
ctx = context.WithValue(context.Background(), grc.UseCacheKey, true)
ctx = context.WithValue(ctx, grc.CacheTTLKey, 10*time.Second)
db.Session(&gorm.Session{Context: ctx}).Find(&users)
```

## 🔧 Available Cache Implementations

### Your Own Implementations (Recommended)

The best approach is to use your preferred cache libraries:

- **go-redis**: Popular Redis client library
- **go-cache**: Thread-safe in-memory cache  
- **BigCache**: Fast, concurrent, evicting in-memory cache
- **FreeCache**: Zero GC overhead cache
- **Memcached**: Official memcached client
- **Database cache**: Use your existing database as cache
- **File cache**: Custom file-based cache

### Reference Implementations (`examples/implementations/`)

These are provided as examples and starting points:

- **MemoryCache**: Production-ready in-memory cache with cleanup
- **SimpleRedisClient**: Redis implementation without go-redis dependency
- **MemcachedCache**: Memcached implementation example
- **FileCache**: File-based persistent cache example

### Integration Examples

See `examples/implementations/popular_libraries.go` for examples of how to wrap:
- go-redis
- BigCache
- Other popular cache libraries

### Create Your Own

Implement any storage backend by satisfying the `CacheClient` interface:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

See `examples/implementations/README.md` for detailed implementation guides.

## 📋 Complete Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/evangwt/grc"
    "github.com/evangwt/grc/examples/implementations"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

type User struct {
    ID   uint
    Name string
}

func main() {
    // Setup database
    db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    
    db.AutoMigrate(&User{})
    
    // Choose your cache implementation
    cacheBackend := implementations.NewMemoryCache() // Reference implementation
    defer cacheBackend.Close()
    
    // Create and register cache
    cache := grc.NewGormCache("user_cache", cacheBackend, grc.CacheConfig{
        TTL:    5 * time.Minute,
        Prefix: "users:",
    })
    
    db.Use(cache)
    
    // Create test data
    db.Create(&User{Name: "Alice"})
    db.Create(&User{Name: "Bob"})
    
    var users []User
    
    // Query with cache
    ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
    
    // First query - cache miss, hits database
    db.Session(&gorm.Session{Context: ctx}).Find(&users)
    log.Printf("First query: %d users", len(users))
    
    // Second query - cache hit, no database query
    db.Session(&gorm.Session{Context: ctx}).Find(&users)
    log.Printf("Second query: %d users (from cache)", len(users))
}
```

## 🎛️ Cache Control

### Enable/Disable Cache

```go
// Use cache with default TTL
ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)

// Do not use cache  
ctx := context.WithValue(context.Background(), grc.UseCacheKey, false)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
// or simply
db.Where("id > ?", 10).Find(&users)
```

### Custom TTL

```go
// Use cache with custom TTL
ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
ctx = context.WithValue(ctx, grc.CacheTTLKey, 10*time.Second)
db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 5).Find(&users)
```

## 🧪 Testing & Development

grc provides comprehensive testing capabilities:

- **Abstract Interface**: Test any cache implementation against the `CacheClient` interface
- **Redis Testing**: Uses `miniredis` for integration testing without external Redis server
- **Reference Implementations**: Use examples for development and testing

Run tests:
```bash
go test ./...
```

## 📚 Examples

For comprehensive examples and implementation guides:

- **`examples/implementations/`** - Reference cache implementations (memory, Memcached, file-based)
- **`example/memory_example.go`** - Complete usage demonstration  

## 🔄 Migration Guide

### From go-redis based solutions:

**Before:**
```go
rdb := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
cache := grc.NewGormCache("cache", grc.NewRedisClient(rdb), config)
```

**After:**
```go
// Option 1: Use your go-redis client directly (RECOMMENDED)
type MyGoRedisCache struct {
    client *redis.Client
}

func (r *MyGoRedisCache) Get(ctx context.Context, key string) (interface{}, error) {
    val, err := r.client.Get(ctx, key).Result()
    if err != nil {
        if err == redis.Nil {
            return nil, grc.ErrCacheMiss
        }
        return nil, err
    }
    return []byte(val), nil
}

func (r *MyGoRedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
    data, err := json.Marshal(value)
    if err != nil {
        return err
    }
    return r.client.Set(ctx, key, data, ttl).Err()
}

myCache := &MyGoRedisCache{client: rdb}
cache := grc.NewGormCache("cache", myCache, config)

// Option 2: Use reference implementation (for testing/development)
redisClient, _ := implementations.NewSimpleRedisClient(implementations.SimpleRedisConfig{
    Addr: "localhost:6379",
})
cache := grc.NewGormCache("cache", redisClient, config)
```

## 📄 License

grc is licensed under the Apache License 2.0. See the [LICENSE](https://github.com/evangwt/grc/blob/main/LICENSE) file for more information.

## 🤝 Contribution

If you have any feedback or suggestions for grc, please feel free to open an issue or a pull request on GitHub. Your contribution is welcome and appreciated! 😊

---

**grc v2**: 简洁优雅的使用方式 (Simple and Elegant Usage) with Clean Abstract Interface Design 🚀

**Why use grc?** Because your cache implementation is YOUR choice. grc provides the elegant interface - you provide the cache that fits your needs.

