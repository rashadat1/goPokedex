package pokecache

import (
	"sync"
	"time"
)

type cacheEntry struct {
	createdAt      time.Time
	val            []byte
}
type Cache struct {
	entryMap       map[string]cacheEntry
	mut            *sync.Mutex     
}

func (c *Cache) Add(key string, val []byte) {
	c.mut.Lock()
	newEntry := cacheEntry{createdAt: time.Now().UTC(), val: val}
	c.entryMap[key] = newEntry
	c.mut.Unlock()
}
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mut.Lock()
	defer c.mut.Unlock()
	retrievedCacheEntry, ok := c.entryMap[key]
	if !ok {
		return nil, false
	}
	return retrievedCacheEntry.val, true
}
func (c *Cache) reapLoop(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case _ = <-ticker.C:
			c.mut.Lock()
			// remove logic here for expired cache entries
			currTimestamp := time.Now().UTC()
			for key, val := range c.entryMap {
				if currTimestamp.Sub(val.createdAt) > interval {
					delete(c.entryMap, key)
				}
			}
			c.mut.Unlock()
		}
	}
}

func NewCache(interval time.Duration) *Cache {
	entries := make(map[string]cacheEntry)
	mutex := sync.Mutex{}
	cache := Cache{entryMap: entries, mut: &mutex}
	
	go cache.reapLoop(interval)
	return &cache
}
