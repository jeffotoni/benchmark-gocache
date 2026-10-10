package main

import (
	"strconv"
	"testing"
	"time"

	bigcache "github.com/allegro/bigcache"
	freecache "github.com/coocood/freecache"
	ristretto "github.com/dgraph-io/ristretto"
	currentgocache "github.com/jeffotoni/gocache"
	gocache "github.com/patrickmn/go-cache"
)

const (
	currentBenchmarkKeyCount = 4096
	currentBenchmarkTTL      = 10 * time.Minute
)

var currentBenchmarkKeys = makeCurrentBenchmarkKeys("benchmark:key:")
var currentBenchmarkMissKeys = makeCurrentBenchmarkKeys("benchmark:miss:")
var currentBenchmarkByteKeys = makeCurrentBenchmarkByteKeys(currentBenchmarkKeys)
var currentBenchmarkMissByteKeys = makeCurrentBenchmarkByteKeys(currentBenchmarkMissKeys)
var currentBenchmarkPayload = []byte("benchmark-value")

func makeCurrentBenchmarkKeys(prefix string) []string {
	keys := make([]string, currentBenchmarkKeyCount)
	for index := range keys {
		keys[index] = prefix + strconv.Itoa(index)
	}
	return keys
}

func makeCurrentBenchmarkByteKeys(keys []string) [][]byte {
	byteKeys := make([][]byte, len(keys))
	for index, key := range keys {
		byteKeys[index] = []byte(key)
	}
	return byteKeys
}

// BenchmarkCurrentSet uses the same key sequence, payload and TTL for every
// implementation. New and NewFast come from the local gocache module.
func BenchmarkCurrentSet(b *testing.B) {
	benchmarkCurrentImplementations(b, func(b *testing.B, name string) {
		switch name {
		case "gocache-New":
			cache := currentgocache.New(currentBenchmarkTTL)
			defer cache.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				cache.Set(currentBenchmarkKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload, currentgocache.DefaultExpiration)
			}
		case "gocache-NewFast":
			cache := currentgocache.NewFast(currentBenchmarkTTL)
			defer cache.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				cache.Set(currentBenchmarkKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload, currentgocache.DefaultExpiration)
			}
		case "go-cache":
			cache := gocache.New(currentBenchmarkTTL, time.Minute)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				cache.Set(currentBenchmarkKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload, gocache.DefaultExpiration)
			}
		case "freecache":
			cache := freecache.NewCache(100 * 1024 * 1024)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if err := cache.Set(currentBenchmarkByteKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload, int(currentBenchmarkTTL/time.Second)); err != nil {
					b.Fatal(err)
				}
			}
		case "ristretto":
			cache := newCurrentRistretto(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if !cache.SetWithTTL(currentBenchmarkKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload, 1, currentBenchmarkTTL) {
					b.Fatal("ristretto rejected Set")
				}
			}
		case "bigcache":
			cache := newCurrentBigcache(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if err := cache.Set(currentBenchmarkKeys[index%currentBenchmarkKeyCount], currentBenchmarkPayload); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

// BenchmarkCurrentSetGet measures a successful Set followed by Get using the
// same key. Ristretto waits for its asynchronous write before reading it.
func BenchmarkCurrentSetGet(b *testing.B) {
	benchmarkCurrentImplementations(b, func(b *testing.B, name string) {
		switch name {
		case "gocache-New":
			cache := currentgocache.New(currentBenchmarkTTL)
			defer cache.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkKeys[index%currentBenchmarkKeyCount]
				cache.Set(key, currentBenchmarkPayload, currentgocache.DefaultExpiration)
				if _, found := cache.Get(key); !found {
					b.Fatal("gocache New Get miss")
				}
			}
		case "gocache-NewFast":
			cache := currentgocache.NewFast(currentBenchmarkTTL)
			defer cache.Close()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkKeys[index%currentBenchmarkKeyCount]
				cache.Set(key, currentBenchmarkPayload, currentgocache.DefaultExpiration)
				if _, found := cache.Get(key); !found {
					b.Fatal("gocache NewFast Get miss")
				}
			}
		case "go-cache":
			cache := gocache.New(currentBenchmarkTTL, time.Minute)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkKeys[index%currentBenchmarkKeyCount]
				cache.Set(key, currentBenchmarkPayload, gocache.DefaultExpiration)
				if _, found := cache.Get(key); !found {
					b.Fatal("go-cache Get miss")
				}
			}
		case "freecache":
			cache := freecache.NewCache(100 * 1024 * 1024)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkByteKeys[index%currentBenchmarkKeyCount]
				if err := cache.Set(key, currentBenchmarkPayload, int(currentBenchmarkTTL/time.Second)); err != nil {
					b.Fatal(err)
				}
				if _, err := cache.Get(key); err != nil {
					b.Fatal(err)
				}
			}
		case "ristretto":
			cache := newCurrentRistretto(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkKeys[index%currentBenchmarkKeyCount]
				if !cache.SetWithTTL(key, currentBenchmarkPayload, 1, currentBenchmarkTTL) {
					b.Fatal("ristretto rejected Set")
				}
				cache.Wait()
				if _, found := cache.Get(key); !found {
					b.Fatal("ristretto Get miss")
				}
			}
		case "bigcache":
			cache := newCurrentBigcache(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				key := currentBenchmarkKeys[index%currentBenchmarkKeyCount]
				if err := cache.Set(key, currentBenchmarkPayload); err != nil {
					b.Fatal(err)
				}
				if _, err := cache.Get(key); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkCurrentGetHit(b *testing.B) {
	benchmarkCurrentImplementations(b, func(b *testing.B, name string) {
		switch name {
		case "gocache-New":
			cache := currentgocache.New(currentBenchmarkTTL)
			defer cache.Close()
			for _, key := range currentBenchmarkKeys {
				cache.Set(key, currentBenchmarkPayload, currentgocache.DefaultExpiration)
			}
			benchmarkCurrentGocacheGetHit(b, cache)
		case "gocache-NewFast":
			cache := currentgocache.NewFast(currentBenchmarkTTL)
			defer cache.Close()
			for _, key := range currentBenchmarkKeys {
				cache.Set(key, currentBenchmarkPayload, currentgocache.DefaultExpiration)
			}
			benchmarkCurrentGocacheGetHit(b, cache)
		case "go-cache":
			cache := gocache.New(currentBenchmarkTTL, time.Minute)
			for _, key := range currentBenchmarkKeys {
				cache.Set(key, currentBenchmarkPayload, gocache.DefaultExpiration)
			}
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, found := cache.Get(currentBenchmarkKeys[index%currentBenchmarkKeyCount]); !found {
					b.Fatal("go-cache Get miss")
				}
			}
		case "freecache":
			cache := freecache.NewCache(100 * 1024 * 1024)
			for index := range currentBenchmarkByteKeys {
				if err := cache.Set(currentBenchmarkByteKeys[index], currentBenchmarkPayload, int(currentBenchmarkTTL/time.Second)); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, err := cache.Get(currentBenchmarkByteKeys[index%currentBenchmarkKeyCount]); err != nil {
					b.Fatal(err)
				}
			}
		case "ristretto":
			cache := newCurrentRistretto(b)
			for _, key := range currentBenchmarkKeys {
				cache.SetWithTTL(key, currentBenchmarkPayload, 1, currentBenchmarkTTL)
			}
			cache.Wait()
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, found := cache.Get(currentBenchmarkKeys[index%currentBenchmarkKeyCount]); !found {
					b.Fatal("ristretto Get miss")
				}
			}
		case "bigcache":
			cache := newCurrentBigcache(b)
			for _, key := range currentBenchmarkKeys {
				if err := cache.Set(key, currentBenchmarkPayload); err != nil {
					b.Fatal(err)
				}
			}
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, err := cache.Get(currentBenchmarkKeys[index%currentBenchmarkKeyCount]); err != nil {
					b.Fatal(err)
				}
			}
		}
	})
}

func BenchmarkCurrentGetMiss(b *testing.B) {
	benchmarkCurrentImplementations(b, func(b *testing.B, name string) {
		switch name {
		case "gocache-New":
			benchmarkCurrentGocacheGetMiss(b, currentgocache.New(currentBenchmarkTTL))
		case "gocache-NewFast":
			benchmarkCurrentGocacheGetMiss(b, currentgocache.NewFast(currentBenchmarkTTL))
		case "go-cache":
			cache := gocache.New(currentBenchmarkTTL, time.Minute)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, found := cache.Get(currentBenchmarkMissKeys[index%currentBenchmarkKeyCount]); found {
					b.Fatal("go-cache returned an unexpected hit")
				}
			}
		case "freecache":
			cache := freecache.NewCache(100 * 1024 * 1024)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, err := cache.Get(currentBenchmarkMissByteKeys[index%currentBenchmarkKeyCount]); err == nil {
					b.Fatal("freecache returned an unexpected hit")
				}
			}
		case "ristretto":
			cache := newCurrentRistretto(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, found := cache.Get(currentBenchmarkMissKeys[index%currentBenchmarkKeyCount]); found {
					b.Fatal("ristretto returned an unexpected hit")
				}
			}
		case "bigcache":
			cache := newCurrentBigcache(b)
			b.ReportAllocs()
			b.ResetTimer()
			for index := 0; index < b.N; index++ {
				if _, err := cache.Get(currentBenchmarkMissKeys[index%currentBenchmarkKeyCount]); err == nil {
					b.Fatal("bigcache returned an unexpected hit")
				}
			}
		}
	})
}

func benchmarkCurrentImplementations(b *testing.B, run func(*testing.B, string)) {
	b.Helper()
	for _, name := range []string{"gocache-New", "gocache-NewFast", "go-cache", "freecache", "ristretto", "bigcache"} {
		b.Run(name, func(b *testing.B) {
			run(b, name)
		})
	}
}

func benchmarkCurrentGocacheGetHit(b *testing.B, cache interface {
	Get(string) (interface{}, bool)
}) {
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, found := cache.Get(currentBenchmarkKeys[index%currentBenchmarkKeyCount]); !found {
			b.Fatal("gocache Get miss")
		}
	}
}

func benchmarkCurrentGocacheGetMiss(b *testing.B, cache interface {
	Get(string) (interface{}, bool)
}) {
	defer func() {
		if closer, ok := cache.(interface{ Close() }); ok {
			closer.Close()
		}
	}()
	b.ReportAllocs()
	b.ResetTimer()
	for index := 0; index < b.N; index++ {
		if _, found := cache.Get(currentBenchmarkMissKeys[index%currentBenchmarkKeyCount]); found {
			b.Fatal("gocache returned an unexpected hit")
		}
	}
}

func newCurrentRistretto(b *testing.B) *ristretto.Cache {
	b.Helper()
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e6,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
	if err != nil {
		b.Fatal(err)
	}
	return cache
}

func newCurrentBigcache(b *testing.B) *bigcache.BigCache {
	b.Helper()
	cache, err := bigcache.NewBigCache(bigcache.DefaultConfig(currentBenchmarkTTL))
	if err != nil {
		b.Fatal(err)
	}
	return cache
}
