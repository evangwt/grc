# grc: Gorm Cache Plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/evangwt/grc)](https://goreportcard.com/report/github.com/evangwt/grc)[![GitHub release](https://img.shields.io/github/release/evangwt/grc.svg)](https://github.com/evangwt/grc/releases/)

grc is a simple and elegant gorm cache plugin with a clean interface design and production-ready default implementations.

## ✨ Features

- **🎯 Simple Interface**: Clean `CacheClient` interface for maximum flexibility
- **🚀 Built-in Implementations**: Production-ready MemoryCache and RedisClient
- **📝 Easy API**: Simple context-based cache control  
- **⚡ High Performance**: Optimized hashing with 27% performance improvement
- **🛡️ Production Ready**: Thread-safe, timeout support, graceful error handling

## 📦 Installation

```bash
go get -u github.com/evangwt/grc/v2
```

## 🚀 Quick Start

### Basic Usage

```go
package main

import (
    "context"
    "time"
    "github.com/evangwt/grc/v2"
    "gorm.io/gorm"
)

func main() {
    // Step 1: Choose a cache implementation
    cache := grc.NewGormCache("my_cache", grc.NewMemoryCache(), grc.CacheConfig{
        TTL:           60 * time.Second,
        Prefix:        "cache:",
        UseSecureHash: false, // Fast FNV hashing (27% faster)
    })
    
    // Step 2: Register with gorm
    db.Use(cache)
    
    // Step 3: Use with context
    ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Find(&users)
}
```

## 🔧 Cache Implementations

### Built-in (Production Ready)

**MemoryCache** - Fast in-memory cache with automatic cleanup:
```go
memCache := grc.NewMemoryCache()
defer memCache.Close()
```

**RedisClient** - Simple Redis client without external dependencies:
```go
redisClient, err := grc.NewRedisClient(grc.RedisConfig{
    Addr:     "localhost:6379",
    Password: "", // optional
    DB:       0,  // optional
})
defer redisClient.Close()
```

### Custom Implementation

Implement the `CacheClient` interface for your own cache:

```go
type CacheClient interface {
    Get(ctx context.Context, key string) (interface{}, error)
    Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}
```

## 🎛️ Cache Control

### Enable/Disable Cache

```go
// Use cache
ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
db.Session(&gorm.Session{Context: ctx}).Find(&users)

// Skip cache  
db.Find(&users) // or set UseCacheKey to false
```

### Custom TTL

```go
ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
ctx = context.WithValue(ctx, grc.CacheTTLKey, 10*time.Second)
db.Session(&gorm.Session{Context: ctx}).Find(&users)
```

## ⚡ Performance

- **FNV Hashing**: 194.7 ns/op (default, 27% faster)
- **SHA256 Hashing**: 265.5 ns/op (secure, collision-resistant)
- **Memory Cache**: Sub-microsecond operations with automatic cleanup
- **Timeout Support**: 5-second default timeout with graceful error handling

## 📋 Complete Example

```go
package main

import (
    "context"
    "log"
    "time"
    
    "github.com/evangwt/grc/v2"
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
    
    // Setup cache
    cache := grc.NewGormCache("user_cache", grc.NewMemoryCache(), grc.CacheConfig{
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

## 📄 License

grc is licensed under the Apache License 2.0. See the [LICENSE](https://github.com/evangwt/grc/blob/main/LICENSE) file for more information.

---

**grc**: Simple and elegant gorm cache plugin with production-ready defaults 🚀

