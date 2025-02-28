package v7

import (
	"sync"
	"time"
)

const (
	NoExpiration      time.Duration = -1
	DefaultExpiration time.Duration = 0
	shardCount                      = 8
)

type Item struct {
	value   any
	expires int64
}

type shard struct {
	mu    sync.RWMutex
	items map[uint32]*Item
}

type Cache struct {
	shards [shardCount]*shard
	ttl    time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{ttl: ttl}
	for i := range c.shards {
		c.shards[i] = &shard{items: make(map[uint32]*Item)}
	}
	return c
}

func hashKey(key string) uint32 {
	var h uint32
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

func (c *Cache) getShard(key string) *shard {
	return c.shards[hashKey(key)%shardCount]
}

func (c *Cache) Set(key string, value any, ttl time.Duration) {
	var expires int64
	if ttl == DefaultExpiration {
		ttl = c.ttl
	}
	if ttl > 0 {
		expires = time.Now().Add(ttl).UnixNano()
	}

	sh := c.getShard(key)
	sh.mu.Lock()
	sh.items[hashKey(key)] = &Item{
		value:   value,
		expires: expires,
	}
	sh.mu.Unlock()
}

func (c *Cache) Get(key string) (any, bool) {
	sh := c.getShard(key)
	sh.mu.RLock()
	item, exists := sh.items[hashKey(key)]
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
	delete(sh.items, hashKey(key))
	sh.mu.Unlock()
}
