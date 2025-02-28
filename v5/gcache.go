package v5

import (
	"hash/fnv"
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
	shardCount                      = 17
)

type Item struct {
	value   interface{}
	expires int64
}

type shard struct {
	mu    sync.RWMutex
	items map[string]*Item
}

type Cache struct {
	shards [shardCount]*shard
	ttl    time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{ttl: ttl}
	for i := range c.shards {
		c.shards[i] = &shard{items: make(map[string]*Item)}
	}
	if ttl > 0 {
		go c.cleanExpired()
	}
	return c
}

func (c *Cache) getShard(key string) *shard {
	hash := fnv.New32a()
	hash.Write([]byte(key))
	index := hash.Sum32() % shardCount
	return c.shards[index]
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	var expires int64
	if ttl == DefaultExpiration {
		ttl = c.ttl
	}
	if ttl > 0 {
		expires = time.Now().Add(ttl).UnixNano()
	}

	sh := c.getShard(key)
	sh.mu.Lock()
	sh.items[key] = &Item{value: value, expires: expires}
	sh.mu.Unlock()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	sh := c.getShard(key)
	sh.mu.RLock()
	item, exists := sh.items[key]
	sh.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		c.Delete(key)
		return nil, false
	}

	return item.value, true
}

func (c *Cache) Delete(key string) {
	sh := c.getShard(key)
	sh.mu.Lock()
	delete(sh.items, key)
	sh.mu.Unlock()
}

func (c *Cache) cleanExpired() {
	ticker := time.NewTicker(c.ttl)
	defer ticker.Stop()

	for range ticker.C {
		for _, sh := range c.shards {
			sh.mu.Lock()
			now := time.Now().UnixNano()
			for key, item := range sh.items {
				if item.expires > 0 && now > item.expires {
					delete(sh.items, key)
				}
			}
			sh.mu.Unlock()
		}
	}
}
