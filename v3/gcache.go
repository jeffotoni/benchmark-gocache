package v3

import (
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
)

type Item struct {
	value   interface{}
	expires int64
}

func (i *Item) isExpired() bool {
	return i.expires > 0 && time.Now().UnixNano() > i.expires
}

type Cache struct {
	mu       sync.RWMutex
	ttl      time.Duration
	items    map[string]*Item
	cleaner  *time.Ticker
	stopChan chan struct{}
}

func New(ttl time.Duration, cleanupInterval time.Duration) *Cache {
	cache := &Cache{
		ttl:      ttl,
		items:    make(map[string]*Item),
		stopChan: make(chan struct{}),
	}

	if cleanupInterval > 0 {
		cache.startCleanup(cleanupInterval)
	}

	return cache
}

func calculateExpiration(ttl time.Duration) int64 {
	if ttl > 0 {
		return time.Now().Add(ttl).UnixNano()
	}
	return 0
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	if ttl == DefaultExpiration {
		ttl = c.ttl
	}
	c.mu.Lock()
	c.items[key] = &Item{
		value:   value,
		expires: calculateExpiration(ttl),
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists || item.isExpired() {
		c.Delete(key) // Remove itens expirados
		return nil, false
	}
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) startCleanup(interval time.Duration) {
	c.cleaner = time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-c.cleaner.C:
				c.Clean()
			case <-c.stopChan:
				c.cleaner.Stop()
				return
			}
		}
	}()
}

func (c *Cache) Clean() {
	c.mu.Lock()
	for key, item := range c.items {
		if item.isExpired() {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}

func (c *Cache) StopCleanup() {
	close(c.stopChan)
}
