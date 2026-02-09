package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// CacheEntry represents a cached item with TTL
type CacheEntry struct {
	Data      interface{} `json:"data"`
	ExpiresAt time.Time   `json:"expires_at"`
}

// Cache provides intelligent caching with TTL and thread-safe operations
type Cache struct {
	dir    string
	locks  map[string]*sync.RWMutex
	lockMu sync.Mutex
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

	return &Cache{
		dir:   cacheDir,
		locks: make(map[string]*sync.RWMutex),
	}, nil
}

// getLock returns a per-key lock for thread-safe access
func (c *Cache) getLock(key string) *sync.RWMutex {
	c.lockMu.Lock()
	defer c.lockMu.Unlock()

	if lock, exists := c.locks[key]; exists {
		return lock
	}

	lock := &sync.RWMutex{}
	c.locks[key] = lock
	return lock
}

// Set stores data in cache with TTL using atomic write (temp file + rename)
func (c *Cache) Set(key string, data interface{}, ttl time.Duration) error {
	lock := c.getLock(key)
	lock.Lock()
	defer lock.Unlock()

	entry := CacheEntry{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl),
	}

	jsonData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("failed to marshal cache entry: %w", err)
	}

	filePath := filepath.Join(c.dir, c.sanitizeKey(key)+".json")

	// Atomic write: write to temp file, then rename
	tempFile := filePath + ".tmp." + c.generateTempSuffix()
	if err := os.WriteFile(tempFile, jsonData, 0644); err != nil {
		return fmt.Errorf("failed to write temp cache file: %w", err)
	}

	if err := os.Rename(tempFile, filePath); err != nil {
		os.Remove(tempFile) // Clean up temp file on error
		return fmt.Errorf("failed to rename cache file: %w", err)
	}

	return nil
}

// sanitizeKey converts a key to a safe filename using SHA256 hash
func (c *Cache) sanitizeKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}

// generateTempSuffix generates a unique suffix for temp files
func (c *Cache) generateTempSuffix() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// Get retrieves data from cache if not expired with thread-safe read lock
func (c *Cache) Get(key string, target interface{}) (bool, error) {
	lock := c.getLock(key)
	lock.RLock()
	defer lock.RUnlock()

	filePath := filepath.Join(c.dir, c.sanitizeKey(key)+".json")

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
		// Expired - need write lock to delete
		lock.RUnlock()
		lock.Lock()
		os.Remove(filePath) // Clean up expired entry
		lock.Unlock()
		lock.RLock()
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
