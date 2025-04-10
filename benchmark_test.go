package main

import (
	// "github.com/cespare/xxhash/v2"

	"io"
	"log"
	"strconv"
	"testing"
	"time"

	freecache "github.com/coocood/freecache"
	gocache "github.com/patrickmn/go-cache"

	bigcache "github.com/allegro/bigcache"
	ristretto "github.com/dgraph-io/ristretto"

	v1 "benchmark-gocache/v1"
	v10 "benchmark-gocache/v10"
	v11 "benchmark-gocache/v11"

	// v2 "benchmark-gocache/v2"
	// v3 "benchmark-gocache/v3"
	// v4 "benchmark-gocache/v4"
	// v5 "benchmark-gocache/v5"
	// v6 "benchmark-gocache/v6"
	// v7 "benchmark-gocache/v7"

	v8 "benchmark-gocache/v8"
	v9 "benchmark-gocache/v9"
)

var cacheV1 = v1.New(10 * time.Minute)

// var cacheV2 = v2.New[string, int](10*time.Minute, 0)
// var cacheV3 = v3.New(10*time.Minute, 1*time.Minute)
// var cacheV4 = v4.New(10 * time.Minute)
// var cacheV5 = v5.New(10 * time.Minute)
// var cacheV6 = v6.New(10 * time.Minute)
// var cacheV7 = v7.New(10 * time.Minute)

var cacheV8 = v8.New(10*time.Minute, 8)
var cacheV9 = v9.New(10 * time.Minute)
var cacheV10 = v10.New(10 * time.Minute)
var cacheV11 = v11.New(10 * time.Minute)

var cacheGoCache = gocache.New(10*time.Second, 1*time.Minute)
var fcacheSize = 100 * 1024 * 1024 // 100MB de cache
var cacheFreeCache = freecache.NewCache(fcacheSize)

var ristrettoCache *ristretto.Cache

func init() {
	var err error
	log.SetOutput(io.Discard) // Silencia logs globais
	ristrettoCache, err = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // número de chaves para estimativa de frequência
		MaxCost:     1 << 30, // custo total em bytes
		BufferItems: 64,      // performance optimization
	})
	if err != nil {
		log.Fatalf("failed to initialize ristretto: %v", err)
	}

}

var bigCacheInstance *bigcache.BigCache

func init() {
	cfg := bigcache.DefaultConfig(10 * time.Minute)
	var err error
	bigCacheInstance, err = bigcache.NewBigCache(cfg)
	if err != nil {
		log.Fatalf("failed to initialize bigcache: %v", err)
	}
}

// func BenchmarkFNV1aShort(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = cacheV11.Xfnv1aHash("example_ke")
// 	}
// }

// func BenchmarkXXHashShort(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = xxhash.Sum64String("example_ke")
// 	}
// }

// func BenchmarkFNV1aLong(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = cacheV11.Xfnv1aHash("example_kex_123445696868098765452323")
// 	}
// }

// func BenchmarkXXHashLong(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		_ = xxhash.Sum64String("example_kex_123445696868098765452323")
// 	}
// }

func BenchmarkGcacheSet1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV1.Set(key, i, time.Duration(time.Minute))
	}
}

func BenchmarkGcacheSetGet1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV1.Set(key, i, 10*time.Minute)
		i, ok := cacheV1.Get(key)
		if !ok {
			log.Printf("Not Found %v", i)
		}
	}
}

// func BenchmarkGcacheSet2(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV2.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet2(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV2.Set(key, i, 10*time.Second)
// 		i, ok := cacheV2.Get(key)
// 		if ok != nil {
// 			log.Printf("Not Found %v i=%v", ok, i)
// 		}
// 	}
// }

// func BenchmarkGcacheSet3(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV3.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet3(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV3.Set(key, i, 10*time.Minute)
// 		i, ok := cacheV3.Get(key)
// 		if !ok {
// 			log.Printf("Not Found %v", i)
// 		}
// 	}
// }

// func BenchmarkGcacheSet4(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV4.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet4(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV4.Set(key, i, time.Duration(10*time.Minute))
// 		i, ok := cacheV4.Get(key)
// 		if !ok {
// 			b.Errorf("Not Found %v", i)
// 		}
// 	}
// }
// func BenchmarkGcacheSet5(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV5.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet5(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV5.Set(key, i, time.Duration(10*time.Minute))
// 		i, ok := cacheV5.Get(key)
// 		if !ok {
// 			b.Errorf("Not Found %v", i)
// 		}
// 	}
// }

// func BenchmarkGcacheSet6(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV6.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet6(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV6.Set(key, i, time.Duration(10*time.Minute))
// 		i, ok := cacheV6.Get(key)
// 		if !ok {
// 			b.Errorf("Not Found %v", i)
// 		}
// 	}
// }

// func BenchmarkGcacheSet7(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV7.Set(key, i, time.Duration(time.Minute))
// 	}
// }

// func BenchmarkGcacheSetGet7(b *testing.B) {
// 	for i := 0; i < b.N; i++ {
// 		key := strconv.Itoa(i)
// 		cacheV7.Set(key, i, time.Duration(10*time.Minute))
// 		i, ok := cacheV7.Get(key)
// 		if !ok {
// 			b.Errorf("Not Found %v", i)
// 		}
// 	}
// }

func BenchmarkGcacheSet8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV8.Set(key, i, time.Duration(time.Minute))
	}
}

func BenchmarkGcacheSetGet8(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV8.Set(key, i, time.Duration(10*time.Minute))
		i, ok := cacheV8.Get(key)
		if !ok {
			b.Errorf("Not found: %v", i)
		}
	}
}

func BenchmarkGcacheSet9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV9.Set(key, i, time.Duration(time.Minute))
	}
}

func BenchmarkGcacheSetGet9(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV9.Set(key, i, time.Duration(10*time.Minute))
		i, ok := cacheV9.Get(key)
		if !ok {
			b.Errorf("Not found: %v", i)
		}
	}
}

// BenchmarkGcacheSet9 measures the performance
// of Set operations using keys longer than 8 characters.
func BenchmarkGcacheSetUnr10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Ensures the key length is greater than 8
		key := "long_keyx_long_keyx_long_keyx_long_keyx_" + strconv.Itoa(i)
		cacheV9.Set(key, i, time.Minute)
	}
}

// BenchmarkGcacheSetGet9 measures the performance
// of Set and Get operations using keys longer than 8 characters.
func BenchmarkGcacheSetGetUnr10(b *testing.B) {
	for i := 0; i < b.N; i++ {
		// Ensures the key length is greater than 8
		key := "long_keyx_long_keyx_long_keyx_long_keyx_" + strconv.Itoa(i)
		cacheV9.Set(key, i, 10*time.Minute)

		val, ok := cacheV9.Get(key)
		if !ok {
			b.Errorf("Key not found: %v", val)
		}
	}
}

// BenchmarkGcacheSetShort11 measures the performance
// of Set and Get operations using keys shorts algorithm Xfnv1aHash
func BenchmarkGcacheSetShort11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV11.Set(key, i, time.Duration(time.Minute))
	}
}

// BenchmarkGcacheSetGetShort11 measures the performance
// of Set and Get operations using keys shorts algorithm Xfnv1aHash
func BenchmarkGcacheSetGetShort11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV11.Set(key, i, time.Duration(10*time.Minute))
		i, ok := cacheV11.Get(key)
		if !ok {
			b.Errorf("Not found: %v", i)
		}
	}
}

// BenchmarkGcacheSetLong11 measures the performance
// of Set and Get operations using keys longs algorithm xxHash
func BenchmarkGcacheSetLong11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := "example_x_12345677889901234567890" + strconv.Itoa(i)
		cacheV11.Set(key, i, time.Duration(time.Minute))
	}
}

// BenchmarkGcacheSetGetLong11 measures the performance
// of Set and Get operations using keys longs algorithm xxHash
func BenchmarkGcacheSetGetLong11(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := "example_x_12345677889901234567890" + strconv.Itoa(i)
		cacheV11.Set(key, i, time.Duration(10*time.Minute))
		i, ok := cacheV11.Get(key)
		if !ok {
			b.Errorf("Not found: %v", i)
		}
	}
}

// BenchmarkGo_cacheSet measures the performance
func BenchmarkGo_cacheSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheGoCache.Set(key, i, 5*time.Second)
	}
}

// BenchmarkGo_cacheSetGet measures the performance
func BenchmarkGo_cacheSetGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheGoCache.Set(key, i, 5*time.Second)
		i, ok := cacheGoCache.Get(key)
		if !ok {
			log.Printf("Not Found %v", i)
		}
	}
}

// BenchmarkFreeCacheSet measures the performance
func BenchmarkFreeCacheSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheFreeCache.Set([]byte(key), []byte(key), 3600)
	}
}

// BenchmarkFreeCacheSetGet measures the performance
func BenchmarkFreeCacheSetGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheFreeCache.Set([]byte(key), []byte(key), 3600)
		got, err := cacheFreeCache.Get([]byte(key))
		if err != nil {
			log.Printf("\nError fetching from FreeCache: %v %v", err, got)
		}
	}
}

func BenchmarkRistrettoSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		ristrettoCache.Set(key, i, 1) // cost arbitrário
	}
}

func BenchmarkRistrettoSetGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		ristrettoCache.Set(key, i, 1)
		if value, found := ristrettoCache.Get(key); !found {
			log.Printf("Not found %v", value)
		}
	}
}

func BenchmarkBigCacheSet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		bigCacheInstance.Set(key, []byte(key))
	}
}

func BenchmarkBigCacheSetGet(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		bigCacheInstance.Set(key, []byte(key))
		_, err := bigCacheInstance.Get(key)
		if err != nil {
			log.Printf("Not found: %v", err)
		}
	}
}
