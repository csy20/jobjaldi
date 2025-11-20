package jobagent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"
)

// This file provides a simplified version for debugging
// Temporarily disable optimizations to isolate issues

// SimpleScrapeMany is a simplified version without circuit breaker/rate limiting for testing
func SimpleScrapeMany(cfgJSON string) (string, error) {
	var cfg scrapeConfig
	if err := json.Unmarshal([]byte(cfgJSON), &cfg); err != nil {
		return "", fmt.Errorf("invalid config json: %w", err)
	}

	if len(cfg.Targets) == 0 {
		return marshalJobs([]Job{})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var wg sync.WaitGroup
	var mu sync.Mutex
	jobs := make([]Job, 0, len(cfg.Targets)*50)
	errs := make([]string, 0, len(cfg.Targets))

	semaphore := make(chan struct{}, 5)

	for _, target := range cfg.Targets {
		wg.Add(1)
		go func(t targetSpec) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			fetcher, company, err := resolveTarget(t.Provider, t.Company)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s/%s: %v", t.Provider, t.Company, err))
				mu.Unlock()
				return
			}

			// Cache check
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

			// Direct fetch without circuit breaker/retry
			targetJobs, err := fetcher(ctx, httpClient, userAgent, company)
			if err != nil {
				mu.Lock()
				errs = append(errs, fmt.Sprintf("%s/%s: %v", t.Provider, t.Company, err))
				mu.Unlock()
				return
			}

			if len(targetJobs) > 0 {
				jobCache.Set(cacheKey, targetJobs)
				mu.Lock()
				jobs = append(jobs, targetJobs...)
				mu.Unlock()
			}
		}(target)
	}

	wg.Wait()

	if len(jobs) == 0 && len(errs) > 0 {
		return "", fmt.Errorf("all targets failed: %s", strings.Join(errs, "; "))
	}

	return marshalJobs(jobs)
}

