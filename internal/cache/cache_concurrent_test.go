package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentCacheReadWrite tests concurrent cache operations
func TestConcurrentCacheReadWrite(t *testing.T) {
	cache, err := NewCache()
	assert.NoError(t, err)
	defer cache.Clear()

	const numGoroutines = 100
	const numOperations = 10

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*numOperations)

	// Concurrent writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("test-key-%d", id%10) // Use same keys to test contention
				data := map[string]interface{}{
					"id":    id,
					"value": fmt.Sprintf("value-%d-%d", id, j),
					"time":  time.Now().Unix(),
				}
				if err := cache.Set(key, data, 1*time.Minute); err != nil {
					errors <- fmt.Errorf("goroutine %d write failed: %w", id, err)
				}
			}
		}(i)
	}

	// Concurrent reads
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numOperations; j++ {
				key := fmt.Sprintf("test-key-%d", id%10)
				var result map[string]interface{}
				_, err := cache.Get(key, &result)
				if err != nil {
					errors <- fmt.Errorf("goroutine %d read failed: %w", id, err)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	var errorList []error
	for err := range errors {
		errorList = append(errorList, err)
	}

	if len(errorList) > 0 {
		t.Errorf("Encountered %d errors during concurrent operations:", len(errorList))
		for i, err := range errorList {
			if i < 10 { // Show first 10 errors
				t.Logf("  Error %d: %v", i+1, err)
			}
		}
		t.FailNow()
	}
}

// TestConcurrentCacheExpiration tests concurrent access with expiring entries
func TestConcurrentCacheExpiration(t *testing.T) {
	cache, err := NewCache()
	assert.NoError(t, err)
	defer cache.Clear()

	const numGoroutines = 50

	var wg sync.WaitGroup

	// Write entries with very short TTL
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("expiring-key-%d", id)
			data := fmt.Sprintf("value-%d", id)
			cache.Set(key, data, 100*time.Millisecond)
		}(i)
	}

	wg.Wait()

	// Wait for expiration
	time.Sleep(200 * time.Millisecond)

	// Concurrent reads of expired entries
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := fmt.Sprintf("expiring-key-%d", id)
			var result string
			found, err := cache.Get(key, &result)
			assert.NoError(t, err)
			assert.False(t, found, "Entry should be expired")
		}(i)
	}

	wg.Wait()
}

// TestConcurrentCacheSameKey tests heavy contention on a single key
func TestConcurrentCacheSameKey(t *testing.T) {
	cache, err := NewCache()
	assert.NoError(t, err)
	defer cache.Clear()

	const numGoroutines = 200
	const key = "hotspot-key"

	var wg sync.WaitGroup
	writeCount := 0
	var writeMu sync.Mutex

	// Many goroutines writing to the same key
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			data := fmt.Sprintf("value-%d", id)
			if err := cache.Set(key, data, 1*time.Minute); err != nil {
				t.Errorf("Write failed: %v", err)
			} else {
				writeMu.Lock()
				writeCount++
				writeMu.Unlock()
			}
		}(i)
	}

	wg.Wait()

	assert.Equal(t, numGoroutines, writeCount, "All writes should succeed")

	// Verify final state is consistent
	var result string
	found, err := cache.Get(key, &result)
	assert.NoError(t, err)
	assert.True(t, found)
	assert.NotEmpty(t, result)
}
