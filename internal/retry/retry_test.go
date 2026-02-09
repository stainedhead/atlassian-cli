package retry

import (
	"context"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsRetryable_RetryableError(t *testing.T) {
	err := &RetryableError{
		Err:        errors.New("service unavailable"),
		StatusCode: 503,
	}
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_NonRetryableError(t *testing.T) {
	err := &NonRetryableError{
		Err:        errors.New("bad request"),
		StatusCode: 400,
	}
	assert.False(t, isRetryable(err))
}

func TestIsRetryable_NetworkTimeout(t *testing.T) {
	// Simulate network timeout error
	err := &timeoutError{message: "operation timed out"}
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_NetworkTemporary(t *testing.T) {
	// Simulate temporary network error
	err := &temporaryError{message: "temporary failure"}
	assert.True(t, isRetryable(err))
}

func TestIsRetryable_RetryableStatusCodes(t *testing.T) {
	retryableCodes := []int{429, 502, 503, 504}

	for _, code := range retryableCodes {
		t.Run(fmt.Sprintf("HTTP_%d", code), func(t *testing.T) {
			err := fmt.Errorf("HTTP %d error", code)
			assert.True(t, isRetryable(err), "Expected HTTP %d to be retryable", code)
		})
	}
}

func TestIsRetryable_NonRetryableStatusCodes(t *testing.T) {
	nonRetryableCodes := []int{400, 401, 403, 404, 409}

	for _, code := range nonRetryableCodes {
		t.Run(fmt.Sprintf("HTTP_%d", code), func(t *testing.T) {
			err := fmt.Errorf("HTTP %d error", code)
			assert.False(t, isRetryable(err), "Expected HTTP %d to be non-retryable", code)
		})
	}
}

func TestIsRetryable_RetryablePatterns(t *testing.T) {
	tests := []struct {
		name  string
		error string
	}{
		{"timeout", "operation timeout"},
		{"connection refused", "connection refused"},
		{"connection reset", "connection reset by peer"},
		{"temporary", "temporary failure"},
		{"connection generic", "failed to establish connection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.error)
			assert.True(t, isRetryable(err))
		})
	}
}

func TestIsRetryable_NonRetryablePatterns(t *testing.T) {
	tests := []struct {
		name  string
		error string
	}{
		{"generic error", "something went wrong"},
		{"parse error", "failed to parse JSON"},
		{"validation error", "invalid input"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := errors.New(tt.error)
			assert.False(t, isRetryable(err))
		})
	}
}

func TestIsRetryable_NilError(t *testing.T) {
	assert.False(t, isRetryable(nil))
}

func TestIsRetryableStatusCode(t *testing.T) {
	assert.True(t, isRetryableStatusCode(429))
	assert.True(t, isRetryableStatusCode(502))
	assert.True(t, isRetryableStatusCode(503))
	assert.True(t, isRetryableStatusCode(504))
	assert.False(t, isRetryableStatusCode(200))
	assert.False(t, isRetryableStatusCode(400))
}

func TestIsNonRetryableStatusCode(t *testing.T) {
	assert.True(t, isNonRetryableStatusCode(400))
	assert.True(t, isNonRetryableStatusCode(401))
	assert.True(t, isNonRetryableStatusCode(403))
	assert.True(t, isNonRetryableStatusCode(404))
	assert.True(t, isNonRetryableStatusCode(409))
	assert.False(t, isNonRetryableStatusCode(200))
	assert.False(t, isNonRetryableStatusCode(503))
}

func TestDo_Success(t *testing.T) {
	config := DefaultConfig()
	attempt := 0

	err := Do(context.Background(), config, func() error {
		attempt++
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 1, attempt)
}

func TestDo_RetryAndSuccess(t *testing.T) {
	config := DefaultConfig()
	attempt := 0

	err := Do(context.Background(), config, func() error {
		attempt++
		if attempt < 2 {
			return &RetryableError{Err: errors.New("temporary failure"), StatusCode: 503}
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 2, attempt)
}

func TestDo_MaxAttemptsExceeded(t *testing.T) {
	config := Config{
		MaxAttempts: 3,
		BaseDelay:   1 * time.Millisecond,
		MaxDelay:    10 * time.Millisecond,
	}
	attempt := 0

	err := Do(context.Background(), config, func() error {
		attempt++
		return &RetryableError{Err: errors.New("persistent failure"), StatusCode: 503}
	})

	assert.Error(t, err)
	assert.Equal(t, 3, attempt)
	assert.Contains(t, err.Error(), "max retry attempts exceeded")
}

func TestDo_NonRetryableError(t *testing.T) {
	config := DefaultConfig()
	attempt := 0

	err := Do(context.Background(), config, func() error {
		attempt++
		return &NonRetryableError{Err: errors.New("bad request"), StatusCode: 400}
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempt)
}

func TestDo_ContextCancellation(t *testing.T) {
	config := Config{
		MaxAttempts: 10,
		BaseDelay:   100 * time.Millisecond,
		MaxDelay:    1 * time.Second,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempt := 0
	err := Do(ctx, config, func() error {
		attempt++
		return &RetryableError{Err: errors.New("temporary failure")}
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, context.DeadlineExceeded)
	// Should have attempted at least once before context cancellation
	assert.GreaterOrEqual(t, attempt, 1)
}

func TestCalculateDelay(t *testing.T) {
	baseDelay := 100 * time.Millisecond
	maxDelay := 5 * time.Second

	// Test exponential backoff
	delay1 := calculateDelay(1, baseDelay, maxDelay)
	delay2 := calculateDelay(2, baseDelay, maxDelay)
	delay3 := calculateDelay(3, baseDelay, maxDelay)

	// Delays should increase exponentially (with jitter)
	// delay1 should be around 100ms (with jitter)
	// delay2 should be around 200ms (with jitter)
	// delay3 should be around 400ms (with jitter)
	assert.Greater(t, delay1, 50*time.Millisecond)
	assert.Less(t, delay1, 250*time.Millisecond)

	assert.Greater(t, delay2, 100*time.Millisecond)
	assert.Less(t, delay2, 500*time.Millisecond)

	assert.Greater(t, delay3, 200*time.Millisecond)
	assert.Less(t, delay3, 1*time.Second)

	// Test max delay cap
	delayLarge := calculateDelay(10, baseDelay, maxDelay)
	assert.LessOrEqual(t, delayLarge, maxDelay*2) // Allow for jitter
}

// Mock network error types for testing
type timeoutError struct {
	message string
}

func (e *timeoutError) Error() string   { return e.message }
func (e *timeoutError) Timeout() bool   { return true }
func (e *timeoutError) Temporary() bool { return false }

var _ net.Error = (*timeoutError)(nil)

type temporaryError struct {
	message string
}

func (e *temporaryError) Error() string   { return e.message }
func (e *temporaryError) Timeout() bool   { return false }
func (e *temporaryError) Temporary() bool { return true }

var _ net.Error = (*temporaryError)(nil)

func TestRetryableError_Error(t *testing.T) {
	err := &RetryableError{
		Err:        errors.New("test error"),
		StatusCode: 503,
	}
	assert.Contains(t, err.Error(), "503")
	assert.Contains(t, err.Error(), "test error")
}

func TestRetryableError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &RetryableError{
		Err:        innerErr,
		StatusCode: 503,
	}
	assert.Equal(t, innerErr, errors.Unwrap(err))
}

func TestNonRetryableError_Error(t *testing.T) {
	err := &NonRetryableError{
		Err:        errors.New("test error"),
		StatusCode: 400,
	}
	assert.Contains(t, err.Error(), "400")
	assert.Contains(t, err.Error(), "test error")
}

func TestNonRetryableError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &NonRetryableError{
		Err:        innerErr,
		StatusCode: 400,
	}
	assert.Equal(t, innerErr, errors.Unwrap(err))
}
