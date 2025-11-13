// Package pokecache is used to cache entries for a set interval of time
package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	entries  map[string]cacheEntry
	interval time.Duration
	mux      *sync.Mutex
}

func NewCache(interval time.Duration) Cache {
	cache := Cache{
		entries:  map[string]cacheEntry{},
		interval: interval,
		mux:      &sync.Mutex{},
	}

	go cache.reapLoop()

	return cache
}

/* Add adds a new entry to the cache */
func (c Cache) Add(key string, val []byte) error {
	c.mux.Lock()
	c.entries[key] = cacheEntry{
		createdAt: time.Now(),
		val:       val,
	}
	c.mux.Unlock()
	return nil
}

/* Get gets an entry from the cache */
func (c Cache) Get(key string) (val []byte, found bool) {
	var output []byte

	c.mux.Lock()
	entry, exists := c.entries[key]
	c.mux.Unlock()

	if !exists {
		return output, false
	}

	output = entry.val
	return output, true
}

/* reapLoop runs every _interval_ and removes events older than the interval */
func (c Cache) reapLoop() {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for t := range ticker.C {
		for key, entry := range c.entries {
			cutoffTime := entry.createdAt.Add(c.interval)
			if t.After(cutoffTime) {
				c.mux.Lock()
				delete(c.entries, key)
				c.mux.Unlock()
			}
		}
	}
}
