package v3

import (
	"testing"
	"time"
)

func TestCache_SetAndGet(t *testing.T) {
	cache := New(5*time.Second, 0) // TTL de 5s sem limpeza automática

	cache.Set("key1", "value1", DefaultExpiration)
	cache.Set("key2", 12345, DefaultExpiration)

	// Verifica se os valores são recuperados corretamente
	if val, found := cache.Get("key1"); !found || val != "value1" {
		t.Errorf("Expected 'value1', got %v", val)
	}

	if val, found := cache.Get("key2"); !found || val != 12345 {
		t.Errorf("Expected 12345, got %v", val)
	}
}

func TestCache_Expiration(t *testing.T) {
	cache := New(1*time.Second, 0) // TTL de 1s sem limpeza automática

	cache.Set("key", "expired_value", DefaultExpiration)
	time.Sleep(2 * time.Second) // Aguarda a expiração

	if val, found := cache.Get("key"); found {
		t.Errorf("Expected key to be expired, but got %v", val)
	}
}

func TestCache_Delete(t *testing.T) {
	cache := New(5*time.Second, 0)

	cache.Set("key", "value", DefaultExpiration)
	cache.Delete("key")

	if _, found := cache.Get("key"); found {
		t.Errorf("Expected key to be deleted, but it was found")
	}
}

func TestCache_Clean(t *testing.T) {
	cache := New(2*time.Second, 0)

	cache.Set("key1", "val1", 1*time.Second) // Expira em 1s
	cache.Set("key2", "val2", 3*time.Second) // Expira em 3s

	time.Sleep(2 * time.Second) // Aguarda a expiração de "key1"
	cache.Clean()

	if _, found := cache.Get("key1"); found {
		t.Errorf("Expected 'key1' to be cleaned up")
	}

	if _, found := cache.Get("key2"); !found {
		t.Errorf("Expected 'key2' to still exist")
	}
}

func TestCache_CleanupRoutine(t *testing.T) {
	cache := New(1*time.Second, 500*time.Millisecond) // TTL de 1s, limpeza a cada 500ms

	cache.Set("key", "value", DefaultExpiration)
	time.Sleep(2 * time.Second) // Aguarda a limpeza automática

	if _, found := cache.Get("key"); found {
		t.Errorf("Expected 'key' to be automatically cleaned")
	}

	cache.StopCleanup() // Para o processo de limpeza
}

func TestCache_Concurrency(t *testing.T) {
	cache := New(5*time.Second, 0)
	totalOps := 1000

	for i := 0; i < totalOps; i++ {
		go cache.Set("key", i, DefaultExpiration)
	}

	time.Sleep(1 * time.Second) // Espera goroutines finalizarem

	val, found := cache.Get("key")
	if !found {
		t.Errorf("Expected key to exist, but it was not found")
	}

	t.Logf("Final value: %v", val) // Apenas para visualização do último valor salvo
}
