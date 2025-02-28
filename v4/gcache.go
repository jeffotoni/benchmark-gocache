package v4

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

type Cache struct {
	items sync.Map
	ttl   time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{
		ttl: ttl,
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

	c.items.Store(key, &Item{
		value:   value,
		expires: expires,
	})
}

func (c *Cache) Get(key string) (interface{}, bool) {
	val, exists := c.items.Load(key)
	if !exists {
		return nil, false
	}

	item := val.(*Item)
	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		c.items.Delete(key)
		return nil, false
	}
	return item.value, true
}

func (c *Cache) Delete(key string) {
	c.items.Delete(key)
}

func (c *Cache) cleanExpired() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		c.items.Range(func(key, value interface{}) bool {
			item := value.(*Item)
			if item.expires > 0 && time.Now().UnixNano() > item.expires {
				c.items.Delete(key)
			}
			return true
		})
	}
}
