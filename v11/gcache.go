package v11

import (
	"github.com/cespare/xxhash/v2"
	"sync"
	"time"
)

const (
	DefaultExpiration time.Duration = 0    // Uses default TTL if not specified
	NoExpiration      time.Duration = -1   // Items with no expiration time
	numShards                       = 8    // Number of shards for concurrent access
	ringSize                        = 4096 // Size of the expiration ring buffer
)

// ringNode represents an entry in the expiration ring buffer.
type ringNode struct {
	key     uint64 // Hashed key
	expires int64  // Expiration timestamp in nanoseconds
}

// shard is a partition of the cache with its own locking mechanism.
type shard struct {
	mu       sync.RWMutex     // Mutex for concurrent access
	items    map[uint64]*Item // Cached items
	ringBuf  []ringNode       // Ring buffer for tracking expiration
	ringHead int              // Current position in the ring buffer
}

// Item represents a single cache entry.
type Item struct {
	value   any   // Stored value
	expires int64 // Expiration timestamp
}

// Cache is a sharded in-memory cache with expiration handling.
type Cache struct {
	shards [numShards]*shard // Array of shards to reduce contention
	ttl    time.Duration     // Default time-to-live for cache entries
}

// New creates a new instance of Cache with a given TTL.
func New(ttl time.Duration) *Cache {
	c := &Cache{ttl: ttl}
	for i := 0; i < numShards; i++ {
		c.shards[i] = &shard{
			items:   make(map[uint64]*Item),
			ringBuf: make([]ringNode, ringSize),
		}
	}
	if ttl > 0 {
		go c.cleanup()
	}
	return c
}

// hashKey computes a hash value for a given string key.
//
// The function selects the hashing algorithm dynamically based on the key length:
// - If the key length is ≤ 10 characters, it uses FNV-1a (fast for short strings).
// - If the key length is > 10 characters, it uses xxHash (optimized for long strings).
//
// This hybrid approach ensures optimal performance by leveraging FNV-1a’s efficiency
// for small keys while taking advantage of xxHash’s superior speed for large keys.
//
// Returns a uint64 hash value.func (c *Cache) hashKey(key string) uint64 {
func (c *Cache) hashKey(key string) uint64 {

	if len(key) <= 10 {
		return c.Xfnv1aHash(key) // Para chaves curtas, FNV-1a
	}
	return xxhash.Sum64String(key) // Para chaves longas, xxHash
}

// hashKey computes a simple hash from the string key using FNV-1a variation.
func (c *Cache) Xfnv1aHash(key string) uint64 {
	var h uint64
	for i := 0; i < len(key); i++ {
		h ^= uint64(key[i])
		h *= 16777619
	}
	return h
}

// getShard selects the shard based on the hash value.
func (c *Cache) getShard(k uint64) *shard {
	return c.shards[k%numShards]
}

// Set inserts a value into the cache with an optional TTL.
func (c *Cache) Set(key string, value any, ttl time.Duration) {
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

// Get retrieves a value from the cache.
// If the item has expired, it is deleted and returns (nil, false).
func (c *Cache) Get(key string) (any, bool) {
	hashed := c.hashKey(key)
	sh := c.getShard(hashed)

	sh.mu.RLock()
	item, exists := sh.items[hashed]
	sh.mu.RUnlock()

	if !exists {
		return nil, false
	}

	if item.expires > 0 && time.Now().UnixNano() > item.expires {
		c.Delete(key) // Remove expired item
		return nil, false
	}

	return item.value, true
}

// Delete removes an item from the cache.
func (c *Cache) Delete(key string) {
	hashed := c.hashKey(key)
	sh := c.getShard(hashed)

	sh.mu.Lock()
	delete(sh.items, hashed)
	sh.mu.Unlock()
}

// cleanup periodically removes expired items from the cache.
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
