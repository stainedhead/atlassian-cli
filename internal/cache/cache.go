package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// CacheEntry represents a cached item with TTL
type CacheEntry struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Cache provides intelligent caching with TTL
type Cache struct {
	dir string
}

// NewCache creates a new cache instance
func NewCache() (*Cache, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	cacheDir := filepath.Join(home, ".atlassian-cli", "cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	return &Cache{dir: cacheDir}, nil
}

// Set stores data in cache with TTL
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) error {
	entry := CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	filePath := filepath.Join(c.dir, key+".json")
	return os.WriteFile(filePath, jsonData, 0644)
}

// Get retrieves data from cache if not expired
func (c *Cache) Get(key string, target interface{}) (bool, error) {
	filePath := filepath.Join(c.dir, key+".json")
	
	data, err := os.ReadFile(filePath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to read cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return false, fmt.Errorf("failed to unmarshal cache entry: %w", err)
	}

	if time.Now().After(entry.ExpiresAt) {
		os.Remove(filePath) // Clean up expired entry
		return false, nil
	}

	entryData, err := json.Marshal(entry.Data)
	if err != nil {
		return false, fmt.Errorf("failed to marshal entry data: %w", err)
	}

	if err := json.Unmarshal(entryData, target); err != nil {
		return false, fmt.Errorf("failed to unmarshal target: %w", err)
	}

	return true, nil
}

// Clear removes all cached entries
func (c *Cache) Clear() error {
	return os.RemoveAll(c.dir)
}