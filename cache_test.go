package grc

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type TestUser struct {
	ID   int
	Name string
}

var (
	db        *gorm.DB
	userCount = 100
)

func init() {
	var err error

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPwd := os.Getenv("DB_PWD")
	dbName := os.Getenv("DB_NAME")
	
	// Skip database setup if environment variables are not set
	if dbHost == "" || dbPort == "" || dbUser == "" || dbName == "" {
		log.Println("Database environment variables not set, tests will be skipped")
		return
	}
	
	dsn := fmt.Sprintf("host='%v' port='%v' user='%v'  password='%v' dbname='%v' sslmode=disable", dbHost, dbPort, dbUser, dbPwd, dbName)

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Printf("Failed to connect to database: %v, tests will be skipped", err)
		return
	}
	db.Migrator().DropTable(TestUser{})

	db.AutoMigrate(TestUser{})

	for i := 0; i < userCount; i++ {
		if err = db.Save(&TestUser{Name: fmt.Sprintf("%X", byte('A'+i))}).Error; err != nil {
			log.Printf("Failed to create test data: %v", err)
			return
		}
	}
}

// TestCache tests the cache plugin functionality using MemoryCache
func TestCache(t *testing.T) {
	// Skip this test if database is not available
	if db == nil {
		t.Skip("Database not available, skipping test")
		return
	}

	var err error

	// Use the new MemoryCache instead of RedisClient
	cache := NewGormCache("my_cache", NewMemoryCache(), CacheConfig{
		TTL:    60 * time.Second,
		Prefix: "cache:",
	})
	err = db.Use(cache)
	assert.NoError(t, err)

	args := []struct {
		UseCache bool
		TTL      time.Duration
		ID       int
	}{
		{
			UseCache: false,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       10,
		},
		{
			UseCache: true,
			ID:       10,
		},
		{
			UseCache: true,
			TTL:      5 * time.Second,
			ID:       5,
		},
		{
			UseCache: true,
			ID:       15,
		},
		{
			UseCache: true,
			TTL:      10 * time.Second,
			ID:       10,
		},
	}

	for _, arg := range args {
		var users []TestUser
		ctx := context.WithValue(context.Background(), UseCacheKey, arg.UseCache)
		if arg.TTL > 0 {
			ctx = context.WithValue(ctx, CacheTTLKey, arg.TTL)
		}

		// query with cache and custom ttl
		err = db.Session(&gorm.Session{Context: ctx}).Where("id > ?", arg.ID).Find(&users).Error
		assert.NoError(t, err)
		assert.Equal(t, userCount-arg.ID, len(users))
	}
}

// BenchmarkCache benchmarks the cache plugin performance using MemoryCache
func BenchmarkCache(b *testing.B) {
	// Skip this benchmark if database is not available
	if db == nil {
		b.Skip("Database not available, skipping benchmark")
		return
	}

	cache := NewGormCache("my_cache", NewMemoryCache(), CacheConfig{
		TTL:    10 * time.Second,
		Prefix: "cache:",
	})
	db.Use(cache)

	var users []TestUser

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		db.Session(&gorm.Session{Context: context.WithValue(context.Background(), UseCacheKey, true)}).Where("id > ?", 10).Find(&users)
	}
}
