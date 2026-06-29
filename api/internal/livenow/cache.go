package livenow

// ⟡ per-channel TTL cache — stops public endpoints hammering the db

import (
	"sync"
	"time"
)

type cacheEntry[T any] struct {
	data   T
	expiry time.Time
}

type cache[T any] struct {
	mu      sync.Mutex
	entries map[string]cacheEntry[T]
	ttl     time.Duration
}

func newCache[T any](ttl time.Duration) *cache[T] {
	return &cache[T]{entries: make(map[string]cacheEntry[T]), ttl: ttl}
}

func (c *cache[T]) get(key string) (T, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.entries[key]
	if !ok || time.Now().After(e.expiry) {
		var zero T
		return zero, false
	}
	return e.data, true
}

func (c *cache[T]) set(key string, val T) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries[key] = cacheEntry[T]{data: val, expiry: time.Now().Add(c.ttl)}
}
