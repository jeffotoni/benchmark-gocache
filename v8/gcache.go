package v8

import (
	"container/heap"
	"hash/fnv"
	"sync"
	"time"
)

type Item struct {
	key     string
	value   interface{}
	expires int64
	index   int // Indica a posição no heap
}

func (i *Item) isExpired() bool {
	return i.expires > 0 && time.Now().UnixNano() > i.expires
}

type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].expires < pq[j].expires
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // Para segurança
	*pq = old[0 : n-1]
	return item
}

type shard struct {
	items map[string]*Item
	mu    sync.RWMutex
	pq    PriorityQueue
}

type Cache struct {
	shards        []*shard
	numShards     int
	defaultTTL    time.Duration
	cleanupTicker *time.Ticker
	stopCleanup   chan struct{}
}

func New(ttl time.Duration, numShards int,
	cleanupInterval ...time.Duration) *Cache {
	nShards := 8 // Padrão
	if numShards > 0 {
		nShards = numShards
	}

	cleanupInt := ttl / 2 // Padrão: metade do TTL
	if len(cleanupInterval) > 0 && cleanupInterval[0] > 0 {
		cleanupInt = cleanupInterval[0]
	}

	c := &Cache{
		numShards:     nShards,
		defaultTTL:    ttl,
		shards:        make([]*shard, nShards),
		cleanupTicker: time.NewTicker(cleanupInt),
		stopCleanup:   make(chan struct{}),
	}

	for i := 0; i < nShards; i++ {
		c.shards[i] = &shard{
			items: make(map[string]*Item),
			pq:    make(PriorityQueue, 0),
		}
		heap.Init(&c.shards[i].pq)
	}

	go c.cleanupLoop()

	return c
}

func (c *Cache) getShard(key string) *shard {
	hash := int(hashKey(key)) // Use uma hash function eficiente
	return c.shards[hash%len(c.shards)]
}

func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	sh := c.getShard(key)
	expires := calculateExpiration(ttl)
	item := &Item{key: key, value: value, expires: expires}

	sh.mu.Lock()
	defer sh.mu.Unlock()

	if oldItem := sh.items[key]; oldItem != nil {
		heap.Remove(&sh.pq, oldItem.index)
	}
	sh.items[key] = item
	heap.Push(&sh.pq, item)
}

func (c *Cache) Get(key string) (interface{}, bool) {
	sh := c.getShard(key)
	sh.mu.RLock()
	defer sh.mu.RUnlock()

	item, exists := sh.items[key]
	if !exists || item.isExpired() {
		return nil, false
	}
	return item.value, true
}

func (c *Cache) Delete(key string) {
	sh := c.getShard(key)
	sh.mu.Lock()
	defer sh.mu.Unlock()

	if item := sh.items[key]; item != nil {
		heap.Remove(&sh.pq, item.index)
		delete(sh.items, key)
	}
}

func (c *Cache) cleanupLoop() {
	for {
		select {
		case <-c.cleanupTicker.C:
			c.cleanup()
		case <-c.stopCleanup:
			return
		}
	}
}

func (c *Cache) cleanup() {
	now := time.Now().UnixNano()
	for _, sh := range c.shards {
		sh.mu.Lock()
		for {
			if len(sh.pq) == 0 {
				break
			}
			min := sh.pq[0]
			if min.expires > now {
				break
			}
			heap.Pop(&sh.pq)
			delete(sh.items, min.key)
		}
		sh.mu.Unlock()
	}
}

func hashKey(key string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(key))
	return h.Sum32()
}

func calculateExpiration(ttl time.Duration) int64 {
	if ttl == 0 {
		return 0
	}
	return time.Now().Add(ttl).UnixNano()
}
