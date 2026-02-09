package audit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestConcurrentAuditLogging tests concurrent audit log writes
func TestConcurrentAuditLogging(t *testing.T) {
	// Create temp directory for test
	tempDir := t.TempDir()

	// Override the logger's directory by creating it directly
	logger := &Logger{}
	logFile := filepath.Join(tempDir, "audit.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	assert.NoError(t, err)
	defer file.Close()

	logger.file = file

	const numGoroutines = 100
	const numLogs = 10

	var wg sync.WaitGroup

	// Concurrent log writes
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < numLogs; j++ {
				user := fmt.Sprintf("user-%d", id)
				command := fmt.Sprintf("command-%d-%d", id, j)
				args := []string{fmt.Sprintf("arg-%d", j)}

				logger.Log(user, command, args, true, nil)

				// Small random delay to increase contention
				if j%3 == 0 {
					time.Sleep(time.Microsecond)
				}
			}
		}(i)
	}

	wg.Wait()

	// Close the file to flush
	file.Close()

	// Verify log file exists and has content
	stat, err := os.Stat(logFile)
	assert.NoError(t, err)
	assert.Greater(t, stat.Size(), int64(0), "Log file should have content")

	// Read and verify log entries
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)

	lines := 0
	for _, b := range content {
		if b == '\n' {
			lines++
		}
	}

	expectedLines := numGoroutines * numLogs
	assert.Equal(t, expectedLines, lines, "Should have correct number of log entries")
}

// TestConcurrentAuditLoggingWithErrors tests concurrent logging with mixed success/error
func TestConcurrentAuditLoggingWithErrors(t *testing.T) {
	tempDir := t.TempDir()

	logger := &Logger{}
	logFile := filepath.Join(tempDir, "audit.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	assert.NoError(t, err)
	defer file.Close()

	logger.file = file

	const numGoroutines = 50

	var wg sync.WaitGroup

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Log success
			logger.Log(fmt.Sprintf("user-%d", id), "success-command", []string{}, true, nil)

			// Log error
			logger.Log(fmt.Sprintf("user-%d", id), "error-command", []string{}, false,
				fmt.Errorf("test error %d", id))
		}(i)
	}

	wg.Wait()
	file.Close()

	// Verify content
	content, err := os.ReadFile(logFile)
	assert.NoError(t, err)
	assert.Greater(t, len(content), 0)
}

// TestConcurrentAuditLoggerClose tests that Close is safe under concurrency
func TestConcurrentAuditLoggerClose(t *testing.T) {
	tempDir := t.TempDir()

	logger := &Logger{}
	logFile := filepath.Join(tempDir, "audit.log")
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	assert.NoError(t, err)

	logger.file = file

	var wg sync.WaitGroup

	// Start logging in background
	stopLogging := make(chan bool)
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-stopLogging:
				return
			default:
				logger.Log("user", "command", []string{}, true, nil)
				time.Sleep(time.Millisecond)
			}
		}
	}()

	// Let it run briefly
	time.Sleep(10 * time.Millisecond)

	// Close while logging is happening
	err = logger.Close()
	close(stopLogging)

	wg.Wait()

	// Close should not panic and should return without error or with a reasonable error
	// (The error might be "file already closed" which is acceptable)
	assert.True(t, err == nil || err.Error() == "close already closed file")
}
