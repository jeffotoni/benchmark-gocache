package v9

import (
	"sync"
	"time"
)

const (
	DefaultExpiration time.Duration = 0
	NoExpiration      time.Duration = -1
	numShards                       = 8
	ringSize                        = 4096
)

type ringNode struct {
	key     uint32
	expires int64
}

type shard struct {
	mu       sync.RWMutex
	items    map[uint32]*Item
	ringBuf  []ringNode
	ringHead int
}

type Item struct {
	value   interface{}
	expires int64
}

type Cache struct {
	shards [numShards]*shard
	ttl    time.Duration
}

func New(ttl time.Duration) *Cache {
	c := &Cache{ttl: ttl}
	for i := 0; i < numShards; i++ {
		c.shards[i] = &shard{
			items:   make(map[uint32]*Item),
			ringBuf: make([]ringNode, ringSize),
		}
	}
	if ttl > 0 {
		go c.cleanup()
	}
	return c
}

func (c *Cache) hashKey(key string) uint32 {
	var h uint32
	for i := 0; i < len(key); i++ {
		h ^= uint32(key[i])
		h *= 16777619
	}
	return h
}

func (c *Cache) getShard(k uint32) *shard {
	return c.shards[k%numShards]
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	var exp int64
	if ttl == DefaultExpiration {
		ttl = c.ttl
	}
	if ttl > 0 {
		exp = time.Now().Add(ttl).UnixNano()
	}

	hashed := c.hashKey(key)
	sh := c.getShard(hashed)

	sh.mu.Lock()
	sh.items[hashed] = &Item{value: value, expires: exp}
	sh.ringBuf[sh.ringHead] = ringNode{key: hashed, expires: exp}
	sh.ringHead = (sh.ringHead + 1) % ringSize
	sh.mu.Unlock()
}

func (c *Cache) Get(key string) (interface{}, bool) {
	hashed := c.hashKey(key)
	sh := c.getShard(hashed)

	sh.mu.RLock()
	item, exists := sh.items[hashed]
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
	hashed := c.hashKey(key)
	sh := c.getShard(hashed)

	sh.mu.Lock()
	delete(sh.items, hashed)
	sh.mu.Unlock()
}

func (c *Cache) cleanup() {
	tick := time.NewTicker(c.ttl / 2)
	defer tick.Stop()

	for range tick.C {
		now := time.Now().UnixNano()
		for _, sh := range c.shards {
			sh.mu.Lock()
			for i := 0; i < ringSize; i++ {
				node := &sh.ringBuf[i]
				if node.expires > 0 && now > node.expires {
					delete(sh.items, node.key)
					node.expires = 0
				}
			}
			sh.mu.Unlock()
		}
	}
}
