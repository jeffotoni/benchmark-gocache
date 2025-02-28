package v8

import (
	"fmt"
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New(time.Minute, 8, 1*time.Second)
	cache.Set("key1", "value1", 10*time.Second)

	val, found := cache.Get("key1")
	if !found {
		t.Errorf("Expected key1 to be found")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}
}

func TestCache_GetNonExistent(t *testing.T) {
	cache := New(time.Minute, 8, 1*time.Second)
	_, found := cache.Get("nonexistent")
	if found {
		t.Errorf("Did not expect to find nonexistent key")
	}
}

func TestCache_TTLExpiration(t *testing.T) {
	cache := New(time.Minute, 8, 1*time.Second)
	cache.Set("key1", "value1", 1*time.Second)

	time.Sleep(2 * time.Second)

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be expired")
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New(time.Minute, 8, 1*time.Second)
	cache.Set("key1", "value1", 10*time.Second)
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be deleted")
	}
}

func TestCache_Cleanup(t *testing.T) {
	cache := New(time.Minute, 8, 500*time.Millisecond)
	cache.Set("key1", "value1", 1*time.Second)
	cache.Set("key2", "value2", 1*time.Second)

	time.Sleep(2 * time.Second)

	_, found := cache.Get("key1")
	if found {
		t.Errorf("Expected key1 to be cleaned up")
	}
	_, found = cache.Get("key2")
	if found {
		t.Errorf("Expected key2 to be cleaned up")
	}
}

func BenchmarkCache_Set(b *testing.B) {
	cache := New(time.Minute, 8, 1*time.Second)
	for i := 0; i < b.N; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i, 10*time.Second)
	}
}

func BenchmarkCache_Get(b *testing.B) {
	cache := New(time.Minute, 8, 1*time.Second)
	for i := 0; i < 100000; i++ {
		cache.Set(fmt.Sprintf("key%d", i), i, 10*time.Second)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.Get(fmt.Sprintf("key%d", i%100000))
	}
}
