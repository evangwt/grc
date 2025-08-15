package implementations

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"time"
	
	"github.com/evangwt/grc"
)

// FileCache demonstrates how to implement grc.CacheClient for file-based caching
type FileCache struct {
	basePath string
	mu       sync.RWMutex
}

type fileCacheItem struct {
	Value  json.RawMessage `json:"value"`
	Expiry time.Time       `json:"expiry"`
}

// NewFileCache creates a new file-based cache client
func NewFileCache(basePath string) (*FileCache, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	fc := &FileCache{basePath: basePath}
	
	// Start cleanup goroutine
	go fc.cleanup()
	
	return fc, nil
}

// Get retrieves a value from the file cache
func (f *FileCache) Get(ctx context.Context, key string) (interface{}, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	
	filename := filepath.Join(f.basePath, key+".cache")
	
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, grc.ErrCacheMiss
		}
		return nil, err
	}
	
	var item fileCacheItem
	if err := json.Unmarshal(data, &item); err != nil {
		return nil, err
	}
	
	// Check if expired
	if time.Now().After(item.Expiry) {
		// Clean up expired file
		os.Remove(filename)
		return nil, grc.ErrCacheMiss
	}
	
	return []byte(item.Value), nil
}

// Set stores a value in the file cache with TTL
func (f *FileCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	
	item := fileCacheItem{
		Value:  json.RawMessage(data),
		Expiry: time.Now().Add(ttl),
	}
	
	fileData, err := json.Marshal(item)
	if err != nil {
		return err
	}
	
	filename := filepath.Join(f.basePath, key+".cache")
	return ioutil.WriteFile(filename, fileData, 0644)
}

// cleanup removes expired cache files
func (f *FileCache) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	
	for {
		select {
		case <-ticker.C:
			f.cleanupExpired()
		}
	}
}

func (f *FileCache) cleanupExpired() {
	f.mu.Lock()
	defer f.mu.Unlock()
	
	files, err := ioutil.ReadDir(f.basePath)
	if err != nil {
		return
	}
	
	now := time.Now()
	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".cache" {
			filename := filepath.Join(f.basePath, file.Name())
			
			data, err := ioutil.ReadFile(filename)
			if err != nil {
				continue
			}
			
			var item fileCacheItem
			if err := json.Unmarshal(data, &item); err != nil {
				continue
			}
			
			if now.After(item.Expiry) {
				os.Remove(filename)
			}
		}
	}
}

// Close cleans up the file cache (optional)
func (f *FileCache) Close() error {
	// Optional: clean up all cache files
	return nil
}