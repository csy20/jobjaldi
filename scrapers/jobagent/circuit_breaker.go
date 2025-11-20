package jobagent

import (
	"sync"
	"time"
)

// CircuitState represents the state of a circuit breaker
type CircuitState int

const (
	StateClosed CircuitState = iota
	StateOpen
	StateHalfOpen
)

// CircuitBreaker implements the circuit breaker pattern
type CircuitBreaker struct {
	mu                sync.RWMutex
	state             CircuitState
	failureCount      int
	successCount      int
	maxFailures       int
	halfOpenMaxSuccess int
	resetTimeout      time.Duration
	lastFailureTime   time.Time
}

// CircuitBreakerConfig holds configuration for circuit breaker
type CircuitBreakerConfig struct {
	MaxFailures       int
	ResetTimeout      time.Duration
	HalfOpenMaxSuccess int
}

// DefaultCircuitBreakerConfig returns default circuit breaker configuration
func DefaultCircuitBreakerConfig() CircuitBreakerConfig {
	return CircuitBreakerConfig{
		MaxFailures:       10, // Increased from 5 to be less aggressive
		ResetTimeout:      15 * time.Second, // Reduced from 30s for faster recovery
		HalfOpenMaxSuccess: 1, // Reduced from 2 for faster recovery
	}
}

// NewCircuitBreaker creates a new circuit breaker with default config
func NewCircuitBreaker() *CircuitBreaker {
	return NewCircuitBreakerWithConfig(DefaultCircuitBreakerConfig())
}

// NewCircuitBreakerWithConfig creates a new circuit breaker with custom config
func NewCircuitBreakerWithConfig(cfg CircuitBreakerConfig) *CircuitBreaker {
	return &CircuitBreaker{
		state:             StateClosed,
		maxFailures:       cfg.MaxFailures,
		resetTimeout:      cfg.ResetTimeout,
		halfOpenMaxSuccess: cfg.HalfOpenMaxSuccess,
	}
}

// Call executes a function through the circuit breaker
func (cb *CircuitBreaker) Call(fn func() error) error {
	if !cb.allowRequest() {
		return &CircuitBreakerError{State: cb.getState(), Message: "circuit breaker is open"}
	}

	err := fn()
	cb.onResult(err == nil)
	return err
}

// allowRequest checks if a request should be allowed
func (cb *CircuitBreaker) allowRequest() bool {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	switch cb.state {
	case StateClosed:
		return true
	case StateOpen:
		// Check if reset timeout has passed
		if time.Since(cb.lastFailureTime) >= cb.resetTimeout {
			cb.mu.RUnlock()
			cb.mu.Lock()
			// Double-check after acquiring write lock
			if cb.state == StateOpen && time.Since(cb.lastFailureTime) >= cb.resetTimeout {
				cb.state = StateHalfOpen
				cb.successCount = 0
			}
			cb.mu.Unlock()
			cb.mu.RLock()
			return cb.state == StateHalfOpen
		}
		return false
	case StateHalfOpen:
		return true
	default:
		return false
	}
}

// onResult updates circuit breaker state based on result
func (cb *CircuitBreaker) onResult(success bool) {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	if success {
		cb.onSuccess()
	} else {
		cb.onFailure()
	}
}

// onSuccess handles successful request
func (cb *CircuitBreaker) onSuccess() {
	switch cb.state {
	case StateClosed:
		// Reset failure count on success
		cb.failureCount = 0
	case StateHalfOpen:
		cb.successCount++
		if cb.successCount >= cb.halfOpenMaxSuccess {
			cb.state = StateClosed
			cb.failureCount = 0
			cb.successCount = 0
		}
	}
}

// onFailure handles failed request
func (cb *CircuitBreaker) onFailure() {
	cb.failureCount++
	cb.lastFailureTime = time.Now()

	switch cb.state {
	case StateClosed:
		if cb.failureCount >= cb.maxFailures {
			cb.state = StateOpen
		}
	case StateHalfOpen:
		// Any failure in half-open state opens the circuit
		cb.state = StateOpen
		cb.successCount = 0
	}
}

// getState returns current circuit breaker state (thread-safe)
func (cb *CircuitBreaker) getState() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

// Reset manually resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
}

// CircuitBreakerError represents a circuit breaker error
type CircuitBreakerError struct {
	State   CircuitState
	Message string
}

func (e *CircuitBreakerError) Error() string {
	return e.Message
}

