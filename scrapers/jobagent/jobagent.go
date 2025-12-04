package jobagent

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/csy/jobagent/jobagent/adapters"
	"golang.org/x/net/http2"
)

const userAgent = "csy-jobboard/1.0 (+personal-use)"

// EnableSimpleMode temporarily disables optimizations for debugging
// Set to true to bypass circuit breaker, rate limiting, and retry logic
var EnableSimpleMode = false // Disabled for production/optimization

var (
	// HTTP transport with HTTP/2 support
	httpTransport = &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
		DisableCompression:  false,
		// Performance optimizations
		MaxConnsPerHost:       20,
		ResponseHeaderTimeout: 10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig: &tls.Config{
			NextProtos: []string{"h2", "http/1.1"}, // Enable HTTP/2
		},
	}

	httpClient = &http.Client{
		Timeout:   15 * time.Second,
		Transport: httpTransport,
	}

	providerRegistry = map[string]adapters.Fetcher{
		"greenhouse": adapters.FetchGreenhouse,
		"lever":      adapters.FetchLever,
	}

	jobCache = NewCache(5 * time.Minute)

	// Buffer pool for JSON marshaling to reduce allocations
	jsonBufferPool = sync.Pool{
		New: func() interface{} {
			return &bytes.Buffer{}
		},
	}

	// Circuit breakers per provider
	circuitBreakers = sync.Map{}

	// Rate limiters per provider
	rateLimiters = NewProviderRateLimiters()

	// Retry configuration
	retryConfig = DefaultRetryConfig()
)

func init() {
	// Enable HTTP/2 support (with error handling)
	if err := http2.ConfigureTransport(httpTransport); err != nil {
		// HTTP/2 not available, fallback to HTTP/1.1
		// This is fine, HTTP/1.1 will be used automatically
	}

	// Initialize circuit breakers for each provider
	for provider := range providerRegistry {
		circuitBreakers.Store(provider, NewCircuitBreaker())
	}
}

// getCircuitBreaker returns circuit breaker for a provider
func getCircuitBreaker(provider string) *CircuitBreaker {
	if cb, ok := circuitBreakers.Load(provider); ok {
		return cb.(*CircuitBreaker)
	}
	// Create new circuit breaker if not exists
	cb := NewCircuitBreaker()
	circuitBreakers.Store(provider, cb)
	return cb
}

// Job re-exports the adapter job structure for gomobile bindings.
type Job = adapters.Job

type scrapeConfig struct {
	Targets []targetSpec `json:"targets"`
}

type targetSpec struct {
	Provider string `json:"provider"`
	Company  string `json:"company"`
}

// ScrapeProvider fetches jobs for a single provider/company pair.
func ScrapeProvider(provider, company string) (string, error) {
	fetcher, company, err := resolveTarget(provider, company)
	if err != nil {
		return "", err
	}

	// Optimized cache key generation using strings.Builder
	var cacheKeyBuilder strings.Builder
	cacheKeyBuilder.Grow(len(provider) + len(company) + 1)
	cacheKeyBuilder.WriteString(provider)
	cacheKeyBuilder.WriteByte(':')
	cacheKeyBuilder.WriteString(company)
	cacheKey := cacheKeyBuilder.String()

	if cachedJobs, found := jobCache.Get(cacheKey); found {
		return marshalJobs(cachedJobs)
	}

	// Apply rate limiting (non-blocking check first)
	// Just check - don't block indefinitely
	if !rateLimiters.Allow(provider) {
		// Small delay instead of blocking wait
		time.Sleep(100 * time.Millisecond)
		// Try once more, then proceed anyway
		rateLimiters.Allow(provider) // Try again but don't block
	}

	// Use circuit breaker and retry logic
	ctx := context.Background()
	cb := getCircuitBreaker(provider)
	var jobs []Job

	err = cb.Call(func() error {
		result, retryErr := RetryWithBackoff(ctx, retryConfig, func() (interface{}, error) {
			return fetcher(ctx, httpClient, userAgent, company)
		})
		
		if retryErr != nil {
			return retryErr
		}
		
		jobs = result.([]Job)
		return nil
	})

	if err != nil {
		return "", err
	}

	// Filter jobs by location
	filteredJobs := filterJobsByLocation(jobs)

	jobCache.Set(cacheKey, filteredJobs)
	return marshalJobs(filteredJobs)
}

// ScrapeMany processes a JSON config and fetches from all targets concurrently.
func ScrapeMany(cfgJSON string) (string, error) {
	// Use simple mode if enabled (for debugging)
	if EnableSimpleMode {
		return SimpleScrapeMany(cfgJSON)
	}

	var cfg scrapeConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}

	if len(cfg.Targets) == 0 {
		return marshalJobs([]Job{})
	}

	// Deduplicate targets
	deduplicatedTargets := deduplicateTargets(cfg.Targets)

	// Group targets by provider for potential batching
	groupedTargets := groupTargetsByProvider(deduplicatedTargets)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	// Pre-allocate with better capacity estimate
	estimatedJobs := len(deduplicatedTargets) * 50
	jobs := make([]Job, 0, estimatedJobs)
	// Pre-allocate error slice
	errs := make([]string, 0, len(deduplicatedTargets))

	semaphore := make(chan struct{}, 5)

	// Process targets grouped by provider
	// Convert map to slice for consistent iteration
	for provider := range groupedTargets {
		group := groupedTargets[provider]
		for _, target := range group {
			wg.Add(1)
			go func(t targetSpec) {
				defer wg.Done()

				semaphore <- struct{}{}
				defer func() { <-semaphore }()

				// Apply rate limiting (non-blocking - just check, don't wait)
				// If rate limited, proceed anyway to avoid blocking
				rateLimiters.Allow(t.Provider)

				fetcher, company, err := resolveTarget(t.Provider, t.Company)
				if err != nil {
					mu.Lock()
					errs = append(errs, fmt.Sprintf("%s/%s: %v", t.Provider, t.Company, err))
					mu.Unlock()
					return
				}

				// Optimized cache key generation
				var cacheKeyBuilder strings.Builder
				cacheKeyBuilder.Grow(len(t.Provider) + len(company) + 1)
				cacheKeyBuilder.WriteString(t.Provider)
				cacheKeyBuilder.WriteByte(':')
				cacheKeyBuilder.WriteString(company)
				cacheKey := cacheKeyBuilder.String()

				if cachedJobs, found := jobCache.Get(cacheKey); found {
					mu.Lock()
					jobs = append(jobs, cachedJobs...)
					mu.Unlock()
					return
				}

				// Use circuit breaker and retry logic
				cb := getCircuitBreaker(t.Provider)
				var targetJobs []Job
				
				// Use circuit breaker and retry logic
				// Note: Circuit breaker will handle open state internally
				err = cb.Call(func() error {
					result, retryErr := RetryWithBackoff(ctx, retryConfig, func() (interface{}, error) {
						return fetcher(ctx, httpClient, userAgent, company)
					})
					
					if retryErr != nil {
						return retryErr
					}
					
					if result == nil {
						return fmt.Errorf("fetcher returned nil result")
					}
					
					var ok bool
					targetJobs, ok = result.([]Job)
					if !ok {
						return fmt.Errorf("fetcher returned invalid type: %T", result)
					}
					
					return nil
				})

				if err != nil {
					// Log error but continue with other targets
					mu.Lock()
					errs = append(errs, fmt.Sprintf("%s/%s: %v", t.Provider, t.Company, err))
					mu.Unlock()
					return
				}

				// Filter targetJobs by location before caching/appending
				filteredJobs := filterJobsByLocation(targetJobs)

				// Only cache and add if we got jobs
				if len(filteredJobs) > 0 {
					jobCache.Set(cacheKey, filteredJobs)
					mu.Lock()
					jobs = append(jobs, filteredJobs...)
					mu.Unlock()
				}
			}(target)
		}
	}

	wg.Wait()

	// Return jobs even if some targets failed (partial success)
	if len(jobs) == 0 {
		if len(errs) > 0 {
			return "", fmt.Errorf("all targets failed: %s", strings.Join(errs, "; "))
		}
		// No errors but no jobs - return empty array
		return marshalJobs([]Job{})
	}

	return marshalJobs(jobs)
}

// deduplicateTargets removes duplicate provider/company pairs
func deduplicateTargets(targets []targetSpec) []targetSpec {
	seen := make(map[string]bool, len(targets))
	result := make([]targetSpec, 0, len(targets))

	for _, target := range targets {
		key := fmt.Sprintf("%s:%s", strings.ToLower(target.Provider), strings.ToLower(target.Company))
		if !seen[key] {
			seen[key] = true
			result = append(result, target)
		}
	}

	return result
}

// groupTargetsByProvider groups targets by provider for potential batching
func groupTargetsByProvider(targets []targetSpec) map[string][]targetSpec {
	groups := make(map[string][]targetSpec)
	
	for _, target := range targets {
		provider := strings.ToLower(target.Provider)
		groups[provider] = append(groups[provider], target)
	}
	
	return groups
}

func resolveTarget(provider, company string) (adapters.Fetcher, string, error) {
	key := normalize(provider)
	fetcher, ok := providerRegistry[key]
	if !ok {
		return nil, "", fmt.Errorf("unsupported provider %q", provider)
	}

	company = strings.TrimSpace(company)
	if company == "" {
		return nil, "", errors.New("company cannot be empty")
	}

	return fetcher, company, nil
}

func marshalJobs(jobs []Job) (string, error) {
	if jobs == nil {
		jobs = make([]Job, 0)
	}

	// Use buffer pool to reduce allocations
	buf := jsonBufferPool.Get().(*bytes.Buffer)
	defer func() {
		buf.Reset()
		jsonBufferPool.Put(buf)
	}()

	encoder := json.NewEncoder(buf)
	encoder.SetEscapeHTML(false) // Faster, no HTML escaping needed
	if err := encoder.Encode(jobs); err != nil {
		return "", err
	}

	// Remove trailing newline added by Encode
	result := buf.String()
	if len(result) > 0 && result[len(result)-1] == '\n' {
		result = result[:len(result)-1]
	}
	return result, nil
}

func normalize(v string) string {
	return strings.ToLower(strings.TrimSpace(v))
}

// ClearCache clears the job cache (useful for testing or forced refresh)
func ClearCache() {
	jobCache.Clear()
}

// CleanExpiredCache removes expired entries from cache
func CleanExpiredCache() {
	jobCache.CleanExpired()
}

// ResetCircuitBreakers resets all circuit breakers (useful for recovery)
func ResetCircuitBreakers() {
	circuitBreakers.Range(func(key, value interface{}) bool {
		if cb, ok := value.(*CircuitBreaker); ok {
			cb.Reset()
		}
		return true
	})
}

// ResetCircuitBreaker resets circuit breaker for a specific provider
func ResetCircuitBreaker(provider string) {
	if cb, ok := circuitBreakers.Load(provider); ok {
		if breaker, ok := cb.(*CircuitBreaker); ok {
			breaker.Reset()
		}
	}
}

// filterJobsByLocation filters jobs based on target locations (primarily India)
func filterJobsByLocation(jobs []Job) []Job {
	if len(jobs) == 0 {
		return jobs
	}

	filtered := make([]Job, 0, len(jobs))
	for _, job := range jobs {
		if isTargetLocation(job.Location) {
			filtered = append(filtered, job)
		}
	}
	return filtered
}

// isTargetLocation checks if a location string matches our target locations (India)
func isTargetLocation(location string) bool {
	loc := strings.ToLower(location)

	// Primary target: India
	// Check for "India" but exclude "Indiana" (US state)
	// Common patterns: "Bangalore, India", "India", "Remote - India"
	if strings.Contains(loc, "india") && !strings.Contains(loc, "indiana") {
		return true
	}

	// Major Indian cities
	indianCities := []string{
		"bangalore", "bengaluru",
		"mumbai", "bombay",
		"delhi", "new delhi", "ncr",
		"hyderabad", "secunderabad",
		"pune",
		"chennai", "madras",
		"kolkata", "calcutta",
		"ahmedabad",
		"gurgaon", "gurugram",
		"noida",
		"chandigarh",
		"jaipur",
		"indore",
		"kochi", "cochin",
		"trivandrum", "thiruvananthapuram",
		"mysore", "mysuru",
		"bhubaneswar",
		"coimbatore",
		"nagpur",
		"lucknow",
		"surat",
		"visakhapatnam", "vizag",
	}

	for _, city := range indianCities {
		// Check for word boundaries or simple contains if unambiguous
		if strings.Contains(loc, city) {
			// Double check if city name matches something in Indiana/USA?
			// Most of these are fairly unique to India, or the "Indiana" check above covers generic "India" issues.
			// "Hyderabad" also exists in Pakistan but user context is India vs US.
			// "Kochi" is unique enough in this context (Japanese Kochi is usually not "Kochi, India").
			// "Salem" (in Tamil Nadu) is also in US (Oregon, MA, etc), but not in our major list yet.
			// Sticking to major cities reduces ambiguity.
			return true
		}
	}

	return false
}
