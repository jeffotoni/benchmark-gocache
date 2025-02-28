package v4

import (
	"sync"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New(10 * time.Minute)

	cache.Set("key1", "value1", DefaultExpiration)
	cache.Set("key2", 12345, DefaultExpiration)

	val, found := cache.Get("key1")
	if !found || val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	val, found = cache.Get("key2")
	if !found || val != 12345 {
		t.Errorf("Expected 12345, got %v", val)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := New(1 * time.Second)

	cache.Set("key", "expired_value", 500*time.Millisecond)
	time.Sleep(1 * time.Second)

	val, found := cache.Get("key")
	if found {
		t.Errorf("Expected expired item but found: %v", val)
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New(10 * time.Minute)

	cache.Set("key", "value", DefaultExpiration)
	cache.Delete("key")

	_, found := cache.Get("key")
	if found {
		t.Errorf("Expected item removed but still found")
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := New(500 * time.Millisecond)

	cache.Set("key1", "val1", 200*time.Millisecond)
	cache.Set("key2", "val2", 700*time.Millisecond)

	time.Sleep(600 * time.Millisecond)

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected 'key1' to be removed, but it still exists")
	}

	_, found = cache.Get("key2")
	if !found {
		t.Errorf("Expected 'key2' to still exist, but was removed prematurely")
	}
}

func TestCache_Concurrency(t *testing.T) {
	cache := New(10 * time.Minute)
	var wg sync.WaitGroup

	totalOps := 1000

	// Escrita concorrente
	wg.Add(totalOps)
	for i := 0; i < totalOps; i++ {
		go func(i int) {
			defer wg.Done()
			cache.Set(string(rune(i)), i, DefaultExpiration)
		}(i)
	}
	wg.Wait()

	// Leitura concorrente
	wg.Add(totalOps)
	for i := 0; i < totalOps; i++ {
		go func(i int) {
			defer wg.Done()
			cache.Get(string(rune(i)))
		}(i)
	}
	wg.Wait()
}
