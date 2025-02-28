[![License](https://img.shields.io/github/license/jeffotoni/benchmark-gocache)](https://github.com/jeffotoni/benchmark-gocache/blob/main/LICENSE)
![GitHub last commit](https://img.shields.io/github/last-commit/jeffotoni/benchmark-gocache)
![GitHub forks](https://img.shields.io/github/forks/jeffotoni/benchmark-gocache?style=social)
![GitHub stars](https://img.shields.io/github/stars/jeffotoni/benchmark-gocache)

# 🚀💕 Benchmarks for Go In-Memory Caches

This repository provides a comprehensive benchmark suite comparing **nine custom cache implementations** from [`jeffotoni/gocache`](https://github.com/jeffotoni/gocache) (versions `v1` through `v9`), plus two well-known open-source libraries:

- [`patrickmn/go-cache`](https://github.com/patrickmn/go-cache)
- [`coocood/freecache`](https://github.com/coocood/freecache)

All tests are run on an **Apple M3 Max** machine (Darwin/arm64) to measure both **1-second** and **3-second** benchmark performance (`-benchtime=1s` and `-benchtime=3s`).

## Why Use an In-Memory Cache?

- **Faster Access**: In-memory caches reduce latency by storing frequently accessed data directly in memory, avoiding repeated database or external service calls.
- **Reduced Load**: Caching lowers the workload on databases and APIs, improving overall system throughput.
- **Quick Expiration**: In-memory caches are best for *ephemeral* data where occasional staleness is tolerable, and items can expire quickly.

## Cache Implementations Tested

We benchmarked nine versions from [`jeffotoni/gocache`](https://github.com/jeffotoni/gocache), each using different concurrency strategies, expiration approaches, and internal data structures. Additionally, we tested:

- **`go-cache`**: [patrickmn/go-cache](https://github.com/patrickmn/go-cache)
- **`freecache`**: [coocood/freecache](https://github.com/coocood/freecache)

Below is a snippet showing how the caches are instantiated in our benchmarking suite:

```go
var cacheV1 = v1.New(10 * time.Minute)
var cacheV2 = v2.New[string, int](10*time.Minute, 0)
var cacheV3 = v3.New(10*time.Minute, 1*time.Minute)
var cacheV4 = v4.New(10 * time.Minute)
var cacheV5 = v5.New(10 * time.Minute)
var cacheV6 = v6.New(10 * time.Minute)
var cacheV7 = v7.New(10 * time.Minute)
var cacheV8 = v8.New(10*time.Minute, 8)
var cacheV9 = v9.New(10 * time.Minute)

// Third-party libraries
var cacheGoCache = gocache.New(10*time.Second, 1*time.Minute)
var fcacheSize = 100 * 1024 * 1024 // 100MB cache
var cacheFreeCache = freecache.NewCache(fcacheSize)
```

Each version of gocache implements different optimizations (locking, sharding, ring buffers, etc.) to analyze performance trade-offs.

Example Benchmark Function

Below is an example benchmark test used for Version 1 (v1), focusing on both Set() and Get():

```go
//The result wiil BenchmarkGcacheSet1(b *testing.B)
func BenchmarkGcacheSet1(b *testing.B) {
	for i := 0; i < b.N; i++ {
		key := strconv.Itoa(i)
		cacheV1.Set(key, i, time.Duration(time.Minute))
	}
}

//The result wiil BenchmarkGcacheSetGet1(b *testing.B)
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
...
```

    Note: Similar benchmark functions are repeated for v2 through v9, plus go-cache and freecache.

## 🚀 1-Second Benchmarks

$ go test -bench=. -benchtime=1s

| **Implementation** | **Set Ops**    | **Set ns/op** | **Set/Get Ops** | **Set/Get ns/op** | **Observations**                      |
|--------------------|----------------|---------------|-----------------|-------------------|---------------------------------------|
| **gocache V1**     | 6,459,714      | 259.4 ns/op   | 5,062,861       | 245.0 ns/op       | Fast reads, decent writes             |
| **gocache V2**     | 6,597,314      | 239.5 ns/op   | 4,175,704       | 280.4 ns/op       | Good write speed, average read        |
| **gocache V3**     | 7,094,665      | 259.7 ns/op   | 4,746,934       | 281.6 ns/op       | Balanced performance                  |
| **gocache V4**     | 4,644,594      | 324.4 ns/op   | 3,571,759       | 330.9 ns/op       | ❌ Slower (sync.Map)                  |
| **gocache V5**     | 6,311,216      | 252.6 ns/op   | 4,714,106       | 278.7 ns/op       | Solid all-around                      |
| **gocache V6**     | 7,532,767      | 262.6 ns/op   | 4,865,896       | 256.2 ns/op       | 🔥 Great concurrency                  |
| **gocache V7**     | 8,026,825      | 222.4 ns/op   | 4,978,083       | 244.3 ns/op       | 🏆 **Best write** (1s), fast reads    |
| **gocache V8**     | 4,708,249      | 309.3 ns/op   | 2,513,566       | 399.7 ns/op       | ❌ Slower overall                     |
| **gocache V9**     | 9,295,434      | 215.9 ns/op   | 5,096,511       | 272.7 ns/op       | 🏆 **Fastest write** (lowest ns/op)   |
| **go-cache**       | 6,463,236      | 291.6 ns/op   | 4,698,109       | 290.7 ns/op       | Solid library, slower than V7/V9      |
| **freecache**      | 5,803,242      | 351.1 ns/op   | 2,183,834       | 469.7 ns/op       | 🚀 Decent writes, poor reads          |

## 🚀 3-Second Benchmarks

| **Implementation** | **Set Ops**     | **Set ns/op** | **Get Ops**     | **Get ns/op** | **Observations**                     |
|--------------------|-----------------|---------------|-----------------|---------------|--------------------------------------|
| **gocache V1**     | 17,176,026      | 338.5 ns/op   | 13,891,083      | 268.6 ns/op   | Fast read, solid write               |
| **gocache V2**     | 16,457,449      | 318.5 ns/op   | 12,379,336      | 304.4 ns/op   | Good write speed, average read       |
| **gocache V3**     | 20,858,042      | 310.8 ns/op   | 14,042,400      | 287.1 ns/op   | Balanced, efficient                  |
| **gocache V4**     | 15,255,268      | 422.4 ns/op   | 8,882,214       | 406.3 ns/op   | ❌ Slow (sync.Map)                   |
| **gocache V5**     | 20,500,326      | 348.9 ns/op   | 12,597,715      | 271.7 ns/op   | Good balance                         |
| **gocache V6**     | 21,767,736      | 341.4 ns/op   | 13,085,462      | 297.3 ns/op   | 🔥 Strong concurrency                |
| **gocache V7**     | 27,229,544      | 252.4 ns/op   | 14,574,768      | 268.6 ns/op   | 🏆 **Best write** (3s)               |
| **gocache V8**     | 15,796,894      | 383.5 ns/op   | 8,927,028       | 408.8 ns/op   | ❌ Slower overall                    |
| **gocache V9**     | 24,809,947      | 252.1 ns/op   | 13,225,228      | 275.7 ns/op   | 🏆 **Very fast write**, good read    |
| **go-cache**       | 15,594,752      | 375.4 ns/op   | 14,289,182      | 269.7 ns/op   | 🚀 Excellent reads, slower writes    |
| **freecache**      | 13,303,050      | 402.3 ns/op   | 8,903,779       | 421.4 ns/op   | ❌ Decent write, slow read           |

## 🏅 Benchmark Icons Guide  

These icons indicate key **performance insights** from our benchmarks:  

- 🏆 **Top Performance** → Best result in a specific category (fastest read/write).  
- ❌ **Underperformance** → Notably slower compared to other implementations.  
- 🔥 **Balanced & Scalable** → Strong concurrency, optimized trade-offs.  
- 🚀 **High Speed** → Impressive performance, but not always the absolute fastest.  

💡 Use these indicators to **quickly identify the strengths and weaknesses** of each cache version!  

## 🚀 Key Highlights  

✅ **Best Write Performance**:  
   - 🏆 **V7 and V9** consistently deliver the **fastest writes** (`lowest ns/op` in Set benchmarks).  
   - **V9** achieves top speeds while maintaining strong read performance.  

✅ **Best Read Performance**:  
   - **V1 and go-cache** often provide **the lowest `ns/op` in Get benchmarks**, making them excellent choices for read-heavy workloads.  
   - **go-cache** remains a strong competitor in retrieval speed.  

⚠️ **Slower Performance Observed**:  
   - ❌ **V4 (`sync.Map`) and V8** struggle in both read and write speeds, making them less suitable for high-performance applications.  
   - **freecache** performs well in **writes** but has significantly **slower read speeds**.  

🔥 **Overall, V7 and V9 stand out as the best-balanced options for both write speed and retrieval performance!**  

## ⚖️ Overall Trade-Offs  

Every cache implementation has **its own strengths and weaknesses**:  

✅ **Optimized for Reads** → Some caches prioritize fast retrieval speeds.  
🚀 **High Write Throughput** → Others are designed to handle massive insertions efficiently.  
🔥 **Balanced Performance** → **V7 and V9** strike a great balance between **read and write speeds**.  

💡 **Choosing the right cache depends on your workload needs!**  

---

## 🤝 Contributing  

Want to **enhance this benchmark suite**? Follow these simple steps:  

1️⃣ **Fork this repo** and add your own cache tests or custom versions.  
2️⃣ **Submit a Pull Request (PR)** with your improvements or questions.  
3️⃣ **Join the discussion** by opening an issue to suggest new features or optimizations.  

Your contributions are always welcome! 🚀✨  

---

## 🔗 Related Projects  

This benchmark compares the following caching solutions:  

✅ [**jeffotoni/gocache**](https://github.com/jeffotoni/gocache) – Custom high-performance cache versions (V1–V9).  
✅ [**patrickmn/go-cache**](https://github.com/patrickmn/go-cache) – Lightweight in-memory cache with expiration.  
✅ [**coocood/freecache**](https://github.com/coocood/freecache) – High-speed cache optimized for low GC overhead.  

📌 If you know another cache worth benchmarking, feel free to suggest it!  

---

## 📜 License  

This project is **open-source** under the **MIT License**.  

💡 Feel free to **fork, modify, and experiment** with these benchmarks in your own applications or libraries.  
🔬 The goal is to **help developers choose the best in-memory cache** for their needs.  

🚀 **Happy benchmarking!**