package retry

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"time"
)

// Config defines retry configuration
type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultConfig returns sensible retry defaults
func DefaultConfig() Config {
	return Config{
		MaxAttempts: 3,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    5 * time.Second,
	}
}

// Do executes a function with exponential backoff retry
func Do(ctx context.Context, config Config, fn func() error) error {
	var lastErr error
	
	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := calculateDelay(attempt, config.BaseDelay, config.MaxDelay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		if err := fn(); err != nil {
			lastErr = err
			if !isRetryable(err) {
				return err
			}
			continue
		}
		
		return nil
	}
	
	return fmt.Errorf("max retry attempts exceeded: %w", lastErr)
}

// calculateDelay computes exponential backoff with jitter
func calculateDelay(attempt int, baseDelay, maxDelay time.Duration) time.Duration {
	delay := time.Duration(float64(baseDelay) * math.Pow(2, float64(attempt-1)))
	if delay > maxDelay {
		delay = maxDelay
	}
	
	// Add jitter (±25%)
	jitter := time.Duration(rand.Float64() * float64(delay) * 0.5)
	return delay + jitter - time.Duration(float64(delay)*0.25)
}

// isRetryable determines if an error should trigger a retry
func isRetryable(err error) bool {
	// Simple heuristic - in real implementation, check for specific error types
	errStr := err.Error()
	return contains(errStr, "timeout") || 
		   contains(errStr, "connection") ||
		   contains(errStr, "temporary")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}