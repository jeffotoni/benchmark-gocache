package v2

import (
	"fmt"
	"sync"
	"time"
)

const (
	NoExpiration   time.Duration = -1
	DefaultExpires time.Duration = 0
)

type Item[V any] struct {
	value   V
	expires int64
}

type cache[K ~string, V any] struct {
	mu         sync.RWMutex
	items      map[K]*Item[V]
	done       chan struct{}
	expTime    time.Duration
	cleanupInt time.Duration
}

type Cache[K ~string, V any] struct {
	*cache[K, V]
}

func newCache[K ~string, V any](expTime, cleanupInt time.Duration, item map[K]*Item[V]) *cache[K, V] {
	return &cache[K, V]{
		items:      item,
		expTime:    expTime,
		cleanupInt: cleanupInt,
		done:       make(chan struct{}),
	}
}

func New[K ~string, V any](expTime, cleanupTime time.Duration) *Cache[K, V] {
	items := make(map[K]*Item[V])
	c := newCache(expTime, cleanupTime, items)

	if cleanupTime > 0 {
		go c.cleanup()
	}

	return &Cache[K, V]{c}
}

func (c *Cache[K, V]) Set(key K, val V, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		return fmt.Errorf("item with key '%v' already exists. Use Update() instead", key)
	}
	return c.add(key, val, d)
}

func (c *Cache[K, V]) SetDefault(key K, val V) error {
	return c.Set(key, val, DefaultExpires)
}

func (c *Cache[K, V]) add(key K, val V, d time.Duration) error {
	exp := int64(0)
	if d == DefaultExpires {
		d = c.expTime
	}
	if d > 0 {
		exp = time.Now().Add(d).UnixNano()
	}

	if str, ok := any(val).(string); ok && str == "" {
		return fmt.Errorf("value of type string cannot be empty")
	}

	c.items[key] = &Item[V]{value: val, expires: exp}
	return nil
}

func (c *Cache[K, V]) Get(key K) (*Item[V], error) {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("item with key '%v' not found", key)
	}

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		c.Delete(key)
		return nil, fmt.Errorf("item with key '%v' expired", key)
	}

	return item, nil
}

func (c *Cache[K, V]) Update(key K, val V, d time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; !exists {
		return fmt.Errorf("item with key '%v' does not exist", key)
	}
	return c.add(key, val, d)
}

func (c *Cache[K, V]) Delete(key K) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[key]; exists {
		delete(c.items, key)
		return nil
	}
	return fmt.Errorf("item with key '%v' does not exist", key)
}

func (c *cache[K, V]) DeleteExpired() {
	now := time.Now().UnixNano()
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, item := range c.items {
		if item.expires > 0 && now > item.expires {
			delete(c.items, k)
		}
	}
}

func (c *Cache[K, V]) Flush() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.items = make(map[K]*Item[V])
}

func (c *Cache[K, V]) List() map[K]*Item[V] {
	c.mu.RLock()
	defer c.mu.RUnlock()

	clone := make(map[K]*Item[V], len(c.items))
	for k, v := range c.items {
		clone[k] = v
	}
	return clone
}

func (c *Cache[K, V]) Count() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

func (c *Cache[K, V]) MapToCache(m map[K]V, d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for k, v := range m {
		c.add(k, v, d)
	}
}

func (c *Cache[K, V]) IsExpired(key K) bool {
	c.mu.RLock()
	item, exists := c.items[key]
	c.mu.RUnlock()

	return exists && item.expires > 0 && time.Now().UnixNano() > item.expires
}

func (c *cache[K, V]) cleanup() {
	ticker := time.NewTicker(c.cleanupInt)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.DeleteExpired()
		case <-c.done:
			return
		}
	}
}
