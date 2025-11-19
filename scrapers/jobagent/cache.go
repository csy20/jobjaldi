package jobagent

import (
	"sync"
	"time"

	"github.com/csy/jobagent/jobagent/adapters"
)

type cacheEntry struct {
	jobs      []adapters.Job
	timestamp time.Time
}

type Cache struct {
	mu      sync.RWMutex
	entries map[string]cacheEntry
	ttl     time.Duration
	stop    chan struct{}
	wg      sync.WaitGroup
}

func NewCache(ttl time.Duration) *Cache {
	c := &Cache{
		entries: make(map[string]cacheEntry),
		ttl:     ttl,
		stop:    make(chan struct{}),
	}
	// Start background cleanup goroutine
	c.wg.Add(1)
	go c.cleanupLoop()
	return c
}

func (c *Cache) Get(key string) ([]adapters.Job, bool) {
	c.mu.RLock()
	entry, exists := c.entries[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Fast path: check expiration without lock (timestamp is immutable)
	if time.Since(entry.timestamp) > c.ttl {
		// Entry expired, remove it (need write lock)
		c.mu.Lock()
		// Double-check after acquiring lock
		if entry, stillExists := c.entries[key]; stillExists && time.Since(entry.timestamp) > c.ttl {
			delete(c.entries, key)
		}
		c.mu.Unlock()
		return nil, false
	}

	return entry.jobs, true
}

func (c *Cache) Set(key string, jobs []adapters.Job) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries[key] = cacheEntry{
		jobs:      jobs,
		timestamp: time.Now(),
	}
}

func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.entries = make(map[string]cacheEntry)
}

func (c *Cache) CleanExpired() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.entries {
		if now.Sub(entry.timestamp) > c.ttl {
			delete(c.entries, key)
		}
	}
}

// cleanupLoop runs periodic cleanup in background
func (c *Cache) cleanupLoop() {
	defer c.wg.Done()
	ticker := time.NewTicker(c.ttl / 2) // Clean up every half TTL
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.CleanExpired()
		case <-c.stop:
			return
		}
	}
}

// Close stops the background cleanup goroutine
func (c *Cache) Close() {
	close(c.stop)
	c.wg.Wait()
}
