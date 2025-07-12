# grc: a simple gorm cache plugin

[![Go Report Card](https://goreportcard.com/badge/github.com/evangwt/grc)](https://goreportcard.com/report/github.com/evangwt/grc)[![GitHub release](https://img.shields.io/github/release/evangwt/grc.svg)](https://github.com/evangwt/grc/releases/)

grc is a gorm plugin that provides a simple and flexible way to cache data with **zero external dependencies**.

## Features

- **Zero external dependencies**: Use built-in memory cache or simple Redis client (no go-redis required)
- **Easy to use**: just add grc as a gorm plugin and use gorm session options to control the cache behavior
- **Flexible storage backends**: Choose from in-memory cache or simple Redis implementation
- **Elegant API**: Simple context-based cache control

## Installation

To use grc, you only need gorm installed:

```bash
go get -u gorm.io/gorm
go get -u github.com/evangwt/grc
```

No external cache dependencies required!

## Usage

### Option 1: Memory Cache (Recommended for development)

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

    // Create cache with in-memory storage (no external dependencies!)
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

### Option 2: Simple Redis Client (No go-redis dependency)

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

    // Create simple Redis client (no external dependencies!)
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

    // Use cache
    var users []User
    ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
    db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 10).Find(&users)
}
```

## Cache Control

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

## Migration from v1 (go-redis based)

If you were using the previous version with go-redis, you can easily migrate:

**Before (v1):**
```go
rdb := redis.NewClient(&redis.Options{
    Addr:     "localhost:6379",
    Password: "password",
})
cache := grc.NewGormCache("my_cache", grc.NewRedisClient(rdb), config)
```

**After (v2):**
```go
// Option 1: Use memory cache (no external dependencies)
cache := grc.NewGormCache("my_cache", grc.NewMemoryCache(), config)

// Option 2: Use simple Redis client (no go-redis dependency)
redisClient, _ := grc.NewSimpleRedisClient(grc.SimpleRedisConfig{
    Addr:     "localhost:6379",
    Password: "password",
})
cache := grc.NewGormCache("my_cache", redisClient, config)
```

For more examples and details, please refer to the [example code](https://github.com/evangwt/grc/blob/main/example/).

## License

grc is licensed under the Apache License 2.0 License. See the [LICENSE](https://github.com/evangwt/grc/blob/main/LICENSE) file for more information.

## Contribution

If you have any feedback or suggestions for grc, please feel free to open an issue or a pull request on GitHub. Your contribution is welcome and appreciated.😊

