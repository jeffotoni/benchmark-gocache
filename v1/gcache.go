package v1

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

type cache struct {
	mu    sync.RWMutex
	ttl   time.Duration
	items map[string]*Item
}

type Cache struct {
	*cache
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		cache: &cache{
			ttl:   ttl,
			items: make(map[string]*Item),
		},
	}

	if ttl > 0 {
		go c.cleanExpired()
	}

	return c
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	var expires int64
	if ttl == DefaultExpiration {
		ttl = c.ttl
	}
	if ttl > 0 {
		expires = time.Now().Add(ttl).UnixNano()
	}

	c.mu.Lock()
	c.items[key] = &Item{
		value:   value,
		expires: expires,
	}
	c.mu.Unlock()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, false
	}

	// Se expirado, remove e retorna false
	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

func (c *Cache) cleanExpired() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.clean()
	}
}

func (c *Cache) clean() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	for key, item := range c.items {
		if item.expires > 0 && now > item.expires {
			delete(c.items, key)
		}
	}
	c.mu.Unlock()
}
