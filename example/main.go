package main

import (
	"context"
	"log"
	"time"

	"github.com/evangwt/grc/v2"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	// Use SQLite for this example (no external dependencies needed)
	db, err := gorm.Open(sqlite.Open("example.db"), &gorm.Config{
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

	// =====================================================
	// SIMPLE API: Built-in implementations ready to use
	// =====================================================
	
	// Use built-in MemoryCache - production ready!
	memoryCache := grc.NewGormCache("memory_cache", grc.NewMemoryCache(), grc.CacheConfig{
		TTL:           60 * time.Second,
		Prefix:        "mem:",
		UseSecureHash: false, // Use fast FNV hashing for better performance
	})

	if err := db.Use(memoryCache); err != nil {
		log.Fatal(err)
	}

	// Option 2: Use built-in RedisClient - production ready!
	// redisClient, err := grc.NewRedisClient(grc.RedisConfig{
	//     Addr:        "localhost:6379",
	//     Password:    "", // optional
	//     DB:          0,  // optional
	//     MaxIdleTime: 5 * time.Minute, // optional, auto-reconnect after idle time
	// })
	// if err != nil {
	//     log.Fatal("Failed to connect to Redis:", err)
	// }
	// defer redisClient.Close()
	// 
	// redisCache := grc.NewGormCache("redis_cache", redisClient, grc.CacheConfig{
	//     TTL:           60 * time.Second,
	//     Prefix:        "redis:",
	//     UseSecureHash: true, // Use secure SHA256 for collision resistance if needed
	// })
	// 
	// if err := db.Use(redisCache); err != nil {
	//     log.Fatal(err)
	// }

	// Create some test data
	for i := 1; i <= 10; i++ {
		db.Create(&User{Name: "User" + string(rune('0'+i))})
	}

	var users []User

	// Simple and elegant usage!
	log.Println("=== Using cache with default TTL ===")
	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, true)}).
		Where("id > ?", 5).Find(&users)
	log.Printf("Found %d users", len(users))

	log.Println("=== Using cache with custom TTL ===")
	ctx := context.WithValue(context.Background(), grc.UseCacheKey, true)
	ctx = context.WithValue(ctx, grc.CacheTTLKey, 10*time.Second)
	db.Session(&gorm.Session{Context: ctx}).Where("id > ?", 3).Find(&users)
	log.Printf("Found %d users with 10s TTL", len(users))

	log.Println("=== Query without cache ===")
	db.Session(&gorm.Session{Context: context.WithValue(context.Background(), grc.UseCacheKey, false)}).
		Where("id > ?", 5).Find(&users)
	log.Printf("Found %d users (no cache)", len(users))

	log.Println("=== Performance comparison example ===")
	
	// Demonstrate the difference between fast and secure hashing
	log.Printf("grc supports both fast FNV hashing and secure SHA256 hashing")
	log.Printf("Fast hashing provides ~27%% better performance for most use cases")
	log.Printf("Secure hashing offers collision resistance for high-security scenarios")
	log.Printf("Built-in implementations available: MemoryCache, RedisClient")

	// Show cache name
	log.Printf("Cache '%s' configured successfully", memoryCache.Name())

	log.Println("Example completed successfully!")
}
