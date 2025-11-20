package jobagent

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"
)

// RetryConfig holds configuration for retry behavior
type RetryConfig struct {
	MaxAttempts      int
	InitialBackoff   time.Duration
	MaxBackoff       time.Duration
	BackoffMultiplier float64
}

// DefaultRetryConfig returns a sensible default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:      3,
		InitialBackoff:   1 * time.Second,
		MaxBackoff:       8 * time.Second,
		BackoffMultiplier: 2.0,
	}
}

// httpStatusError interface for checking status codes
type httpStatusError interface {
	StatusCode() int
	Error() string
}

// RetryableError indicates if an error should trigger a retry
func isRetryableError(err error, statusCode int) bool {
	if err == nil {
		return false
	}

	// Check if error has status code
	var httpErr httpStatusError
	if errors.As(err, &httpErr) {
		statusCode = httpErr.StatusCode()
	}

	// Network errors are retryable
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() || netErr.Temporary() {
			return true
		}
	}

	// DNS errors are retryable
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	// 5xx server errors are retryable
	if statusCode >= 500 && statusCode < 600 {
		return true
	}

	// 429 Too Many Requests is retryable
	if statusCode == http.StatusTooManyRequests {
		return true
	}

	// Context cancellation is not retryable
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	return false
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() (interface{}, error)

// RetryWithBackoff executes a function with exponential backoff retry logic
func RetryWithBackoff(ctx context.Context, cfg RetryConfig, fn RetryableFunc) (interface{}, error) {
	var lastErr error
	backoff := cfg.InitialBackoff

	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		// Check context cancellation before each attempt
		if ctx.Err() != nil {
			return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}

		lastErr = err

		// Don't retry on last attempt
		if attempt == cfg.MaxAttempts-1 {
			break
		}

		// Check if error is retryable
		statusCode := 0
		var httpErr httpStatusError
		if errors.As(err, &httpErr) {
			statusCode = httpErr.StatusCode()
		}

		if !isRetryableError(err, statusCode) {
			return nil, err
		}

		// Wait with exponential backoff
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context cancelled during backoff: %w", ctx.Err())
		case <-time.After(backoff):
		}

		// Calculate next backoff
		backoff = time.Duration(float64(backoff) * cfg.BackoffMultiplier)
		if backoff > cfg.MaxBackoff {
			backoff = cfg.MaxBackoff
		}
	}

	return nil, fmt.Errorf("max retry attempts (%d) exceeded: %w", cfg.MaxAttempts, lastErr)
}

// httpError wraps HTTP errors with status code
type httpError struct {
	statusCode int
	message    string
}

func (e *httpError) Error() string {
	return fmt.Sprintf("HTTP %d: %s", e.statusCode, e.message)
}

// NewHTTPError creates a new HTTP error
func NewHTTPError(statusCode int, message string) error {
	return &httpError{statusCode: statusCode, message: message}
}

