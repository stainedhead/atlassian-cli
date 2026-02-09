package retry

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strings"
	"time"
)

// Config defines retry configuration
type Config struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// RetryableError wraps an error that should trigger a retry
type RetryableError struct {
	Err        error
	StatusCode int
}

func (e *RetryableError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("retryable error (HTTP %d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("retryable error: %v", e.Err)
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// NonRetryableError wraps an error that should NOT be retried
type NonRetryableError struct {
	Err        error
	StatusCode int
}

func (e *NonRetryableError) Error() string {
	if e.StatusCode > 0 {
		return fmt.Sprintf("non-retryable error (HTTP %d): %v", e.StatusCode, e.Err)
	}
	return fmt.Sprintf("non-retryable error: %v", e.Err)
}

func (e *NonRetryableError) Unwrap() error {
	return e.Err
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
	if err == nil {
		return false
	}

	// Check for explicit non-retryable error wrapper
	var nonRetryable *NonRetryableError
	if errors.As(err, &nonRetryable) {
		return false
	}

	// Check for explicit retryable error wrapper
	var retryable *RetryableError
	if errors.As(err, &retryable) {
		return true
	}

	// Check for network errors (timeout, temporary)
	var netErr net.Error
	if errors.As(err, &netErr) {
		return netErr.Timeout() || netErr.Temporary()
	}

	// Check for HTTP status code errors
	// We need to check the error message for status codes
	// This is a fallback for when HTTP status codes are embedded in error strings
	errStr := err.Error()

	// Check for retryable HTTP status codes
	for _, code := range []int{429, 502, 503, 504} {
		if strings.Contains(errStr, fmt.Sprintf("%d", code)) ||
			strings.Contains(errStr, http.StatusText(code)) {
			return true
		}
	}

	// Check for non-retryable HTTP status codes
	for _, code := range []int{400, 401, 403, 404, 409} {
		if strings.Contains(errStr, fmt.Sprintf("%d", code)) ||
			strings.Contains(errStr, http.StatusText(code)) {
			return false
		}
	}

	// Check for common retryable error patterns using proper substring matching
	retryablePatterns := []string{"timeout", "connection", "temporary", "connection refused", "connection reset"}
	for _, pattern := range retryablePatterns {
		if strings.Contains(strings.ToLower(errStr), pattern) {
			return true
		}
	}

	// Default to non-retryable
	return false
}

// Helper function for HTTP status code classification
func isRetryableStatusCode(statusCode int) bool {
	return statusCode == 429 || statusCode == 502 || statusCode == 503 || statusCode == 504
}

func isNonRetryableStatusCode(statusCode int) bool {
	return statusCode == 400 || statusCode == 401 || statusCode == 403 || statusCode == 404 || statusCode == 409
}
