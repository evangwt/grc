package grc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"gorm.io/gorm/callbacks"
	"log"
	"time"

	"gorm.io/gorm"
)

var (
	// Context keys with proper typing for better type safety
	UseCacheKey = &contextKey{"UseCache"}
	CacheTTLKey = &contextKey{"CacheTTL"}
	// ErrCacheMiss is returned when a cache key is not found
	ErrCacheMiss = errors.New("cache miss")
	// ErrCacheTimeout is returned when a cache operation times out
	ErrCacheTimeout = errors.New("cache operation timeout")
)

// contextKey provides type safety for context keys
type contextKey struct {
	name string
}

func (c *contextKey) String() string {
	return "grc context key " + c.name
}

// GormCache is a cache plugin for gorm
type GormCache struct {
	name   string
	client CacheClient
	config CacheConfig
}

// CacheClient is an interface for cache operations
type CacheClient interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
}

// CacheConfig is a struct for cache options
type CacheConfig struct {
	TTL           time.Duration // cache expiration time
	Prefix        string        // cache key prefix
	UseSecureHash bool          // use SHA256 instead of FNV (slower but collision-resistant)
}

// NewGormCache returns a new GormCache instance
func NewGormCache(name string, client CacheClient, config CacheConfig) *GormCache {
	return &GormCache{
		name:   name,
		client: client,
		config: config,
	}
}

// Name returns the plugin name
func (g *GormCache) Name() string {
	return g.name
}

// Initialize initializes the plugin
func (g *GormCache) Initialize(db *gorm.DB) error {
	return db.Callback().Query().Replace("gorm:query", g.queryCallback)
}

// queryCallback is a callback function for query operations
func (g *GormCache) queryCallback(db *gorm.DB) {
	if db.Error != nil {
		return
	}

	enableCache := g.enableCache(db)

	// build query sql
	callbacks.BuildQuerySQL(db)
	if db.DryRun || db.Error != nil {
		return
	}

	// Handle caching logic
	if enableCache {
		key := g.cacheKey(db)

		// Try to load from cache first
		hit, err := g.loadCache(db, key)
		if err != nil {
			// Log cache error but don't fail the query
			if !errors.Is(err, ErrCacheTimeout) {
				log.Printf("load cache failed: %v", err)
			}
		} else if hit {
			// Cache hit - return early
			return
		}

		// Cache miss - execute query and cache result
		g.queryDB(db)
		
		// Only cache if query was successful
		if db.Error == nil {
			if err = g.setCache(db, key); err != nil && !errors.Is(err, ErrCacheTimeout) {
				log.Printf("set cache failed: %v", err)
			}
		}
	} else {
		// No caching - execute query directly
		g.queryDB(db)
	}
}

func (g *GormCache) enableCache(db *gorm.DB) bool {
	ctx := db.Statement.Context

	// check if use cache
	useCache, ok := ctx.Value(UseCacheKey).(bool)
	if !ok || !useCache {
		return false // do not use cache, skip this callback
	}
	return true
}

func (g *GormCache) cacheKey(db *gorm.DB) string {
	sql := db.Dialector.Explain(db.Statement.SQL.String(), db.Statement.Vars...)
	
	var hash string
	if g.config.UseSecureHash {
		// Use SHA256 for collision resistance (slower)
		h := sha256.Sum256([]byte(sql))
		hash = hex.EncodeToString(h[:])
	} else {
		// Use FNV-1a for speed (faster, adequate for most use cases)
		h := fnv.New64a()
		h.Write([]byte(sql))
		hash = fmt.Sprintf("%x", h.Sum64())
	}
	
	key := g.config.Prefix + hash
	//log.Printf("key: %v, sql: %v", key, sql)
	return key
}

func (g *GormCache) loadCache(db *gorm.DB, key string) (bool, error) {
	// Add timeout context for cache operations
	ctx := db.Statement.Context
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 5*time.Second {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	value, err := g.client.Get(ctx, key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return false, nil // Cache miss is not an error
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return false, ErrCacheTimeout
		}
		return false, err
	}

	if value == nil {
		return false, nil
	}

	// cache hit, scan value to destination
	if err = json.Unmarshal(value.([]byte), &db.Statement.Dest); err != nil {
		return false, fmt.Errorf("failed to unmarshal cached data: %w", err)
	}
	db.RowsAffected = int64(db.Statement.ReflectValue.Len())
	return true, nil
}

func (g *GormCache) setCache(db *gorm.DB, key string) error {
	ctx := db.Statement.Context

	// get cache ttl from context or config
	ttl, ok := ctx.Value(CacheTTLKey).(time.Duration)
	if !ok {
		ttl = g.config.TTL // use default ttl
	}
	//log.Printf("ttl: %v", ttl)

	// Add timeout context for cache operations
	if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > 5*time.Second {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
	}

	// set value to cache with ttl
	err := g.client.Set(ctx, key, db.Statement.Dest, ttl)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		return ErrCacheTimeout
	}
	return err
}

func (g *GormCache) queryDB(db *gorm.DB) {
	rows, err := db.Statement.ConnPool.QueryContext(db.Statement.Context, db.Statement.SQL.String(), db.Statement.Vars...)
	if err != nil {
		db.AddError(err)
		return
	}
	defer func() {
		db.AddError(rows.Close())
	}()
	gorm.Scan(rows, db, 0)
}
