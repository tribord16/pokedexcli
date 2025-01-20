package pokecache

import (
	"fmt"
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt time.Time
	val       []byte
}

type Cache struct {
	cacheMap map[string]cacheEntry
	mu       sync.Mutex
	interval time.Duration
}

func (c *Cache) Add(key string, valeur []byte) {

	if key == "" || len(valeur) == 0 {
		fmt.Printf("valeur non ajoute %s, %v", key, valeur)
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cacheMap[key] = cacheEntry{
		createdAt: time.Now(),
		val:       valeur,
	}
}

func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.Lock()

	defer c.mu.Unlock()

	cEntry, exists := c.cacheMap[key]
	if !exists {
		return nil, false
	}
	return cEntry.val, true
}

func (c *Cache) reapLoop() {
	for {
		time.Sleep(c.interval)

		c.mu.Lock()

		now := time.Now()

		for key, entry := range c.cacheMap {
			if now.Sub(entry.createdAt) < c.interval {
				delete(c.cacheMap, key)
			}
		}
		c.mu.Unlock()
	}
}

func NewCache(interval time.Duration) *Cache {
	return &Cache{
		cacheMap: make(map[string]cacheEntry),
		mu:       sync.Mutex{},
		interval: interval,
	}

}
