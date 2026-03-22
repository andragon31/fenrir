package validator

import (
	"sync"
	"time"
)

type CacheValidator struct {
	cache map[string]*CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

type CacheEntry struct {
	Value     *PackageInfo
	ExpiresAt time.Time
}

func NewCacheValidator() *CacheValidator {
	return &CacheValidator{
		cache: make(map[string]*CacheEntry),
		ttl:   time.Hour,
	}
}

func (c *CacheValidator) Get(key string) *PackageInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.cache[key]
	if !exists {
		return nil
	}

	if time.Now().After(entry.ExpiresAt) {
		return nil
	}

	return entry.Value
}

func (c *CacheValidator) Set(key string, value *PackageInfo) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache[key] = &CacheEntry{
		Value:     value,
		ExpiresAt: time.Now().Add(c.ttl),
	}
}

func (c *CacheValidator) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.cache, key)
}

func (c *CacheValidator) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.cache = make(map[string]*CacheEntry)
}

func (c *CacheValidator) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return len(c.cache)
}

func (c *CacheValidator) Cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for key, entry := range c.cache {
		if now.After(entry.ExpiresAt) {
			delete(c.cache, key)
		}
	}
}
