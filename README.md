# grc: gorm redis based cache

grc is a gorm plugin that provides a simple and flexible way to cache data using redis.

## Features

- Easy to use: just add grc as a gorm plugin and use gorm session options to control the cache behavior.
- Flexible to customize: you can configure the cache prefix, ttl, and redis client according to your needs.

## Installation

To use grc, you need to have gorm and go-redis installed. You can install them using go get:

```bash
go get -u gorm.io/gorm
go get -u github.com/go-redis/redis/v8
```

Then you can install grc using go get:

```bash
go get -u github.com/evangwt/grc
```

## Usage

To use grc, you need to create a gorm cache instance with a redis client and a cache config, and then add it as a gorm plugin. For example:

```go
package main

import (
        "github.com/evangwt/grc"
        "github.com/go-redis/redis/v8"
        "gorm.io/driver/postgres"
        "gorm.io/gorm"
)

func main() {
        // connect to postgres database
        dsn := "host='0.0.0.0' port='5432' user='evan' dbname='cache_test' password='' sslmode=disable TimeZone=Asia/Shanghai"
        db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{})

        // connect to redis database
        rdb := redis.NewClient(&redis.Options{
                Addr:     "0.0.0.0:6379",
                Password: "123456",
        })

        // create a gorm cache instance with a redis client and a cache config
        cache := grc.NewGormCache("my_cache", grc.NewRedisClient(rdb), grc.CacheConfig{
                TTL:    60 * time.Second,
                Prefix: "cache:",
        })

        // add the gorm cache instance as a gorm plugin
        if err := db.Use(cache); err != nil {
                log.Fatal(err)
        }

        // now you can use gorm session options to control the cache behavior
}
```

To enable or disable the cache for a query, you can use the `grc.UseCacheKey` context value with a boolean value. For example:

```go
// use cache with default ttl
db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, true)}).
                Where("id > ?", 10).Find(&users)

// do not use cache
db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, false)}).
                Where("id > ?", 10).Find(&users)
```

To set a custom ttl for a query, you can use the `grc.CacheTTLKey` context value with a time.Duration value. For example:

```go
// use cache with custom ttl
db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), grc.UseCacheKey, true), grc.CacheTTLKey, 10*time.Second)}).
                Where("id > ?", 5).Find(&users)
```

For more examples and details, please refer to the [example code](https://github.com/evangwt/grc/blob/main/example/main.go).

## License

grc is licensed under the Apache License 2.0 License. See the [LICENSE](https://github.com/evangwt/grc/blob/main/LICENSE) file for more information.

## Contribution

If you have any feedback or suggestions for grc, please feel free to open an issue or a pull request on GitHub. Your contribution is welcome and appreciated.😊

