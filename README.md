# grc: a simple gorm cache plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/evangwt/grc)](https://goreportcard.com/report/github.com/evangwt/grc)[![GitHub release](https://img.shields.io/github/release/evangwt/grc.svg)](https://github.com/evangwt/grc/releases/)

grc is a gorm plugin that provides a **简洁优雅的使用方式** (simple and elegant usage) for data caching with **zero external dependencies**.

## ✨ Features

- **🚀 Zero External Dependencies**: Use built-in memory cache or simple Redis client (no go-redis required)
- **🎯 Elegant Abstract Interface**: Clean `CacheClient` interface allows seamless switching between storage backends
- **💾 Flexible Storage Backends**: Choose from in-memory cache or simple Redis implementation based on your needs
- **📝 Simple Context-Based API**: Control cache behavior through gorm session context
- **🧪 Comprehensive Testing**: Full test coverage with miniredis integration for reliable Redis testing
- **⚡ Production Ready**: Thread-safe implementations suitable for high-concurrency environments
- **🔄 Easy Migration**: Smooth upgrade path from go-redis based solutions

## 🏗️ Architecture

grc implements a clean abstract interface design with the `CacheClient` interface:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

This elegant abstraction allows you to:
- **Switch storage backends** seamlessly without changing your application code
- **Test with different backends** (memory for unit tests, Redis for integration tests)
- **Extend functionality** by implementing custom storage backends
- **Maintain consistency** across different deployment environments

## 📦 Installation

To use grc, you only need gorm installed:

```bash
go get -u gorm.io/gorm
go get -u github.com/evangwt/grc
```

**No external cache dependencies required!** 🎉

## 🚀 Quick Start

### Option 1: Memory Cache (Recommended for Development & Testing)

Perfect for development, testing, and applications that don't require persistent caching:

```go
package main

import (
    "context"
    "time"
    
    "github.com/evangwt/grc"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
)

func main() {
    // Connect to your database
    db, _ := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})

    // Create cache with in-memory storage (zero dependencies!)
    cache := grc.NewGormCache("my_cache", grc.NewMemoryCache(), grc.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "cache:",
    })

    // Add the cache plugin
    if err := db.Use(cache); err != nil {
        log.Fatal(err)
    }

    // Use cache - simple and elegant!
    var users []User
    ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
}
```

### Option 2: Simple Redis Client (Production Ready)

For production environments requiring persistent caching and distributed systems:

```go
package main

import (
    "context"
    "time"
    
    "github.com/evangwt/grc"
    "gorm.io/gorm"
)

func main() {
    // Connect to your database (omitted for brevity)
    db, _ := gorm.Open(...)

    // Create simple Redis client (no go-redis dependency!)
    redisClient, err := grc.NewSimpleRedisClient(grc.SimpleRedisConfig{
        Addr:     "localhost:6379",
        Password: "", // optional
        DB:       0,  // optional
    })
    if err != nil {
        log.Fatal("Failed to connect to Redis:", err)
    }
    defer redisClient.Close()

    // Create cache with Redis storage
    cache := grc.NewGormCache("redis_cache", redisClient, grc.CacheConfig{
        TTL:    60 * time.Second,
        Prefix: "redis:",
    })

    // Add the cache plugin
    if err := db.Use(cache); err != nil {
        log.Fatal(err)
    }

    // Same elegant API regardless of storage backend!
    var users []User
    ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
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

## 🔄 Migration from v1 (go-redis based)

If you were using the previous version with go-redis, you can easily migrate:

**Before (v1):**
```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "password",
})
cache := grc.NewGormCache("my_cache", grc.NewRedisClient(rdb), config)
```

**After (v2 - Zero Dependencies & Elegant Interface):**
```go
// Option 1: Use memory cache (perfect for development/testing)
cache := grc.NewGormCache("my_cache", grc.NewMemoryCache(), config)

// Option 2: Use simple Redis client (production ready, no go-redis dependency)
redisClient, _ := grc.NewSimpleRedisClient(grc.SimpleRedisConfig{
    Addr:     "localhost:6379",
    Password: "password",
})
cache := grc.NewGormCache("my_cache", redisClient, config)
```

**Benefits of Migration:**
- ✅ **Simplified Dependencies**: Eliminate go-redis dependency entirely
- ✅ **Flexible Development**: Use memory cache for testing, Redis for production
- ✅ **Same Elegant API**: No changes to your caching logic required
- ✅ **Better Testing**: Built-in memory cache perfect for unit tests
- ✅ **Production Ready**: SimpleRedisClient optimized for production workloads

## 🧪 Testing & Development

grc provides comprehensive testing capabilities:

- **Memory Cache**: Perfect for unit tests with zero setup required
- **Redis Testing**: Uses `miniredis` for integration testing without external Redis server
- **Interface Testing**: Comprehensive test coverage ensuring both storage backends behave consistently

Run tests:
```bash
go test ./...
```

## 📚 Examples

For comprehensive examples demonstrating both storage backends and advanced usage patterns, please refer to the [example code](https://github.com/evangwt/grc/blob/main/example/).

**Available Examples:**
- `example/main.go` - Complete example showing both memory and Redis usage
- `example/memory_example.go` - Zero-dependency memory cache demonstration

## 📄 License

grc is licensed under the Apache License 2.0 License. See the [LICENSE](https://github.com/evangwt/grc/blob/main/LICENSE) file for more information.

## 🤝 Contribution

If you have any feedback or suggestions for grc, please feel free to open an issue or a pull request on GitHub. Your contribution is welcome and appreciated! 😊

---

**grc v2**: 简洁优雅的使用方式 (Simple and Elegant Usage) with Zero Dependencies 🚀

