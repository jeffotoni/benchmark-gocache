package v9

import (
	"sync"
	"testing"
	"time"
)

// TestCache_SetAndGet verifies that values are correctly
// stored and retrieved from the cache.
func TestCache_SetAndGet(t *testing.T) {
	cache := New(10 * time.Minute)

	// Insert test values with default expiration
	cache.Set("key1", "value1", DefaultExpiration)
	cache.Set("key2", 12345, DefaultExpiration)

	// Retrieve and verify key1
	val, found := cache.Get("key1")
	if !found || val.(string) != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	// Retrieve and verify key2
	val, found = cache.Get("key2")
	if !found || val.(int) != 12345 {
		t.Errorf("Expected 12345, got %v", val)
	}
}

// TestCache_Expiration ensures that expired cache entries
// are removed correctly.
func TestCache_Expiration(t *testing.T) {
	cache := New(1 * time.Second)

	// Set an entry with a short expiration time
	cache.Set("key", "expired_value", 500*time.Millisecond)
	time.Sleep(1 * time.Second) // Wait for expiration

	// Attempt to retrieve the expired entry
	val, found := cache.Get("key")
	if found {
		t.Errorf("Expected expired item, but found: %v", val)
	}
}

// TestCache_Delete verifies that an entry is properly
// removed from the cache.
func TestCache_Delete(t *testing.T) {
	cache := New(10 * time.Minute)

	// Insert and delete an entry
	cache.Set("key", "value", DefaultExpiration)
	cache.Delete("key")

	// Ensure the entry no longer exists
	_, found := cache.Get("key")
	if found {
		t.Errorf("Expected item to be removed, but still found")
	}
}

// TestCache_Cleanup checks that expired items are
// automatically removed during cleanup.
func TestCache_Cleanup(t *testing.T) {
	cache := New(500 * time.Millisecond)

	// Set two keys with different expiration times
	cache.Set("key1", "val1", 200*time.Millisecond)
	cache.Set("key2", "val2", 700*time.Millisecond)

	time.Sleep(600 * time.Millisecond) // Wait for cleanup cycle

	// Verify that 'key1' is expired and removed
	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected 'key1' to be removed, but still exists")
	}

	// Verify that 'key2' still exists
	_, found = cache.Get("key2")
	if !found {
		t.Errorf("Expected 'key2' to still exist, but it was removed early")
	}
}

// TestCache_Concurrency tests cache behavior under concurrent
// read and write operations.
func TestCache_Concurrency(t *testing.T) {
	cache := New(10 * time.Minute)
	var wg sync.WaitGroup

	totalOps := 1000 // Number of concurrent operations

	// Perform concurrent writes
	wg.Add(totalOps)
	for i := 0; i < totalOps; i++ {
		go func(i int) {
			defer wg.Done()
			cache.Set(string(rune(i)), i, DefaultExpiration)
		}(i)
	}
	wg.Wait() // Wait for all writes to complete

	// Perform concurrent reads
	wg.Add(totalOps)
	for i := 0; i < totalOps; i++ {
		go func(i int) {
			defer wg.Done()
			cache.Get(string(rune(i)))
		}(i)
	}
	wg.Wait() // Wait for all reads to complete
}

// TestCacheExpiration test removal and if expired
func TestCacheExpiration(t *testing.T) {
	cache := New(10 * time.Minute)

	key := "test_key"
	value := "test_value"

	// Adds an item with expiration in 1 millisecond
	cache.Set(key, value, 1*time.Millisecond)

	// Please wait for the item to expire.
	time.Sleep(2 * time.Millisecond)

	// Now when calling Get it should remove the item and return false
	result, exists := cache.Get(key)

	if exists {
		t.Errorf("I expected the item to be expired and removed, but it still exists %v", result)
	}
}

// TestNew_DefaultExpiration test Default TTL
func TestNew_DefaultExpiration(t *testing.T) {
	c := New() // Without TTL

	if c == nil {
		t.Fatal("Expected cache instance, got nil")
	}

	if c.ttl != DefaultExpiration {
		t.Errorf("Expected default TTL to be %v, got %v", DefaultExpiration, c.ttl)
	}
}

// TestNew_WithCustomTTL test with TTL
func TestNew_WithCustomTTL(t *testing.T) {
	ttl := 100 * time.Millisecond
	c := New(ttl) // With TTL

	if c == nil {
		t.Fatal("Expected cache instance, got nil")
	}

	if c.ttl != ttl {
		t.Errorf("Expected TTL to be %v, got %v", ttl, c.ttl)
	}
}
