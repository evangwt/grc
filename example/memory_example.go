package main

import (
	"context"
	"log"
	"time"

	"github.com/evangwt/grc"
	"github.com/evangwt/grc/examples/implementations"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Use SQLite for this example (no external dependencies needed)
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: logger.New(
			log.Default(),
			logger.Config{
				LogLevel: logger.Info,
				Colorful: true,
			},
		),
	})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// User is a sample model
	type User struct {
		ID   int
		Name string
	}

	db.AutoMigrate(User{})

	// Create a GormCache with reference MemoryCache implementation - no external dependencies!
	cache := grc.NewGormCache("my_cache", implementations.NewMemoryCache(), grc.CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})

	if err := db.Use(cache); err != nil {
		log.Fatal(err)
	}

	// Create some test data
	for i := 1; i <= 10; i++ {
		db.Create(&User{Name: "User" + string(rune('0'+i))})
	}

	var users []User

	// Use cache with default TTL - simple and elegant!
	log.Println("Querying with cache (first time - cache miss):")
	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, true)}).
		Where("id > ?", 5).Find(&users)
	log.Printf("Found %d users", len(users))

	// Query again - this time it will hit the cache
	log.Println("Querying with cache (second time - cache hit):")
	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, true)}).
		Where("id > ?", 5).Find(&users)
	log.Printf("Found %d users (from cache)", len(users))

	// Use cache with custom TTL
	log.Println("Querying with custom TTL:")
	ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
	ctx = context.WithValue(ctx, grc.CacheTTLKey, 10*time.Second)
	db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 3).Find(&users)
	log.Printf("Found %d users with 10s TTL", len(users))

	// Query without cache
	log.Println("Querying without cache:")
	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, false)}).
		Where("id > ?", 5).Find(&users)
	log.Printf("Found %d users (no cache)", len(users))

	log.Println("Example completed successfully!")
}