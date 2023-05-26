package main

import (
	"context"
	"github.com/evangwt/grc"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"time"
)

func main() {
	dsn := "host='0.0.0.0' port='5432' user='evan' dbname='cache_test' password='' sslmode=disable TimeZone=Asia/Shanghai"
	db, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.Default(),
			logger.Config{
				LogLevel: logger.Info,
				Colorful: true,
			},
		),
	})

	// User is a sample model
	type User struct {
		ID   int
		Name string
	}

	db.AutoMigrate(User{})

	rdb := redis.NewClient(&redis.Options{
		Addr:     "0.0.0.0:6379",
		Password: "123456",
	})

	cache := grc.NewGormCache("my_cache", grc.NewRedisClient(rdb), grc.CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})

	if err := db.Use(cache); err != nil {
		log.Fatal(err)
	}

	var users []User
	/*
		// mock data
		for i := 0; i < 100; i++ {
			db.Save(&User{Name: fmt.Sprintf("%X", byte('A'+i))})
		}
	*/

	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, true)}).
		Where("id > ?", 10).Find(&users) // use cache with default ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), grc.UseCacheKey, true), grc.CacheTTLKey, 10*time.Second)}).
		Where("id > ?", 5).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), grc.UseCacheKey, true), grc.CacheTTLKey, 20*time.Second)}).
		Where("id > ?", 5).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, false)}).
		Where("id > ?", 10).Find(&users) // do not use cache
	log.Printf("users: %#v", users)

	db.Session(&gorm.Session{Context: context.WithValue(context.WithValue(context.Background(), grc.UseCacheKey, true), grc.CacheTTLKey, 10*time.Second)}).
		Where("id > ?", 10).Find(&users) // use cache with custom ttl
	log.Printf("users: %#v", users)
}
