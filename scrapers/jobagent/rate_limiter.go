package jobagent

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu         sync.Mutex
	tokens     float64
	maxTokens  float64
	refillRate float64 // tokens per second
	lastRefill time.Time
}

// RateLimiterConfig holds configuration for rate limiter
type RateLimiterConfig struct {
	Rate  float64 // requests per second
	Burst int     // maximum burst capacity
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	if rate <= 0 {
		rate = 10.0 // default: 10 requests per second
	}
	if burst <= 0 {
		burst = int(rate) // default burst equals rate
	}

	return &RateLimiter{
		tokens:     float64(burst),
		maxTokens:  float64(burst),
		refillRate: rate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed and consumes a token if so
func (rl *RateLimiter) Allow() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()

	// Refill tokens based on elapsed time
	rl.tokens += elapsed * rl.refillRate
	if rl.tokens > rl.maxTokens {
		rl.tokens = rl.maxTokens
	}
	rl.lastRefill = now

	// Check if we have tokens available
	if rl.tokens >= 1.0 {
		rl.tokens -= 1.0
		return true
	}

	return false
}

// WaitForToken blocks until a token is available (with max wait time)
func (rl *RateLimiter) WaitForToken() {
	maxWait := 5 * time.Second // Maximum wait time
	start := time.Now()

	for !rl.Allow() {
		// Check if we've waited too long
		if time.Since(start) > maxWait {
			// Force allow after max wait to prevent indefinite blocking
			return
		}

		// Calculate wait time
		rl.mu.Lock()
		neededTokens := 1.0 - rl.tokens
		if neededTokens <= 0 {
			rl.mu.Unlock()
			return // Should have been allowed
		}
		waitTime := time.Duration(neededTokens/rl.refillRate*1000) * time.Millisecond
		rl.mu.Unlock()

		// Cap wait time to prevent long waits
		if waitTime > 1*time.Second {
			waitTime = 1 * time.Second
		}

		if waitTime > 0 {
			time.Sleep(waitTime)
		} else {
			time.Sleep(10 * time.Millisecond) // small sleep to avoid busy loop
		}
	}
}

// ProviderRateLimiters manages rate limiters per provider
type ProviderRateLimiters struct {
	mu        sync.RWMutex
	limiters  map[string]*RateLimiter
	defaultRL *RateLimiter
}

// NewProviderRateLimiters creates a new provider rate limiters manager
func NewProviderRateLimiters() *ProviderRateLimiters {
	prl := &ProviderRateLimiters{
		limiters: make(map[string]*RateLimiter),
	}

	// Set provider-specific rate limits
	prl.limiters["greenhouse"] = NewRateLimiter(10.0, 20) // 10 req/s, burst 20
	prl.limiters["lever"] = NewRateLimiter(5.0, 10)       // 5 req/s, burst 10 (HTML scraping is heavier)

	// Default rate limiter
	prl.defaultRL = NewRateLimiter(5.0, 10)

	return prl
}

// GetLimiter returns the rate limiter for a provider
func (prl *ProviderRateLimiters) GetLimiter(provider string) *RateLimiter {
	prl.mu.RLock()
	limiter, exists := prl.limiters[provider]
	prl.mu.RUnlock()

	if !exists {
		return prl.defaultRL
	}

	return limiter
}

// Allow checks if a request is allowed for a provider
func (prl *ProviderRateLimiters) Allow(provider string) bool {
	return prl.GetLimiter(provider).Allow()
}

// WaitForToken waits until a request is allowed for a provider
func (prl *ProviderRateLimiters) WaitForToken(provider string) {
	prl.GetLimiter(provider).WaitForToken()
}
