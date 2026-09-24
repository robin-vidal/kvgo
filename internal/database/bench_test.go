package database

import (
	"fmt"
	"hash/fnv"
	"math/rand/v2"
	"runtime"
	"sync"
	"testing"
	"unsafe"

	"github.com/robin-vidal/kvgo/internal/config"
)

const (
	cacheLineSize = 64

	benchKeyspace = 1 << 20

	benchWorkingSet = 1 << 16
)

var (
	benchWorkerCounts = []int{1, 2, 4, 8, 16, 32, 64}
	benchShardCounts  = []int{1, 2, 4, 8, 16, 32, 64, 128, 256}
)

var benchKeys = func() []string {
	keys := make([]string, benchKeyspace)
	for i := range keys {
		keys[i] = fmt.Sprintf("key:%d", i)
	}
	return keys
}()

func workerKeys(seed int) []string {
	r := rand.New(rand.NewPCG(uint64(seed), 0x9E3779B97F4A7C15))
	keys := make([]string, benchWorkingSet)
	for i := range keys {
		keys[i] = benchKeys[r.IntN(len(benchKeys))]
	}
	return keys
}

func runWorkers(b *testing.B, workers int, op func(key string)) {
	b.Helper()

	sets := make([][]string, workers)
	for w := range sets {
		sets[w] = workerKeys(w)
	}

	per := b.N / workers
	if per == 0 {
		per = 1
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(keys []string) {
			defer wg.Done()
			for i := 0; i < per; i++ {
				op(keys[i&(benchWorkingSet-1)])
			}
		}(sets[w])
	}
	wg.Wait()

	b.StopTimer()
}

func BenchmarkDatabaseSet(b *testing.B) {
	for _, shards := range benchShardCounts {
		for _, workers := range benchWorkerCounts {
			b.Run(fmt.Sprintf("shards=%d/workers=%d", shards, workers), func(b *testing.B) {
				db, err := New(&config.Config{ShardAmount: shards})
				if err != nil {
					b.Fatal(err)
				}
				runWorkers(b, workers, func(key string) {
					db.Set(key, "value")
				})
			})
		}
	}
}

type shardCore struct {
	mu   sync.RWMutex
	data map[string]string
}

type shardPadded struct {
	shardCore
	_ [cacheLineSize - unsafe.Sizeof(shardCore{})%cacheLineSize]byte
}

type storeUnpadded struct{ shards []shardCore }

func newStoreUnpadded(n int) *storeUnpadded {
	s := &storeUnpadded{shards: make([]shardCore, n)}
	for i := range s.shards {
		s.shards[i].data = make(map[string]string)
	}
	return s
}

func (s *storeUnpadded) Set(key, value string) {
	shard := &s.shards[getShard(key, len(s.shards))]
	shard.mu.Lock()
	shard.data[key] = value
	shard.mu.Unlock()
}

type storePadded struct{ shards []shardPadded }

func newStorePadded(n int) *storePadded {
	s := &storePadded{shards: make([]shardPadded, n)}
	for i := range s.shards {
		s.shards[i].data = make(map[string]string)
	}
	return s
}

func (s *storePadded) Set(key, value string) {
	shard := &s.shards[getShard(key, len(s.shards))]
	shard.mu.Lock()
	shard.data[key] = value
	shard.mu.Unlock()
}

func BenchmarkShardPadding(b *testing.B) {
	b.Logf("shardCore=%dB shardPadded=%dB cacheLine=%dB GOMAXPROCS=%d",
		unsafe.Sizeof(shardCore{}), unsafe.Sizeof(shardPadded{}),
		cacheLineSize, runtime.GOMAXPROCS(0))

	for _, shards := range benchShardCounts {
		for _, workers := range benchWorkerCounts {
			b.Run(fmt.Sprintf("padding=off/shards=%d/workers=%d", shards, workers), func(b *testing.B) {
				s := newStoreUnpadded(shards)
				runWorkers(b, workers, func(key string) { s.Set(key, "value") })
			})
			b.Run(fmt.Sprintf("padding=on/shards=%d/workers=%d", shards, workers), func(b *testing.B) {
				s := newStorePadded(shards)
				runWorkers(b, workers, func(key string) { s.Set(key, "value") })
			})
		}
	}
}

type storeSingleLock struct {
	mu   sync.RWMutex
	data map[string]string
}

func (s *storeSingleLock) Set(key, value string) {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
}

func BenchmarkSingleMutex(b *testing.B) {
	for _, workers := range benchWorkerCounts {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			s := &storeSingleLock{data: make(map[string]string)}
			runWorkers(b, workers, func(key string) { s.Set(key, "value") })
		})
	}
}

func BenchmarkSyncMap(b *testing.B) {
	for _, workers := range benchWorkerCounts {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			var m sync.Map
			runWorkers(b, workers, func(key string) { m.Store(key, "value") })
		})
	}
}

func BenchmarkGetShard(b *testing.B) {
	b.ReportAllocs()
	keys := workerKeys(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getShard(keys[i&(benchWorkingSet-1)], 8)
	}
}

// Valid only when shardAmount is a power of two.
func getShardMask(key string, shardAmount int) int {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return int(hasher.Sum64() & uint64(shardAmount-1))
}

func BenchmarkGetShardMask(b *testing.B) {
	b.ReportAllocs()
	keys := workerKeys(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getShardMask(keys[i&(benchWorkingSet-1)], 8)
	}
}

// Fresh keys per worker: measures growth, not overwrite. Matters for sync.Map,
// where an overwrite is a CAS but an insert takes the mutex.
func runInsertWorkers(b *testing.B, workers int, op func(key string)) {
	b.Helper()

	const maxKeys = 4 << 20

	per := b.N / workers
	if per == 0 {
		per = 1
	}
	if per*workers > maxKeys {
		b.Skipf("insert benchmark holds one string per op; cap is %d keys, rerun with -benchtime %dx", maxKeys, maxKeys/workers)
	}

	sets := make([][]string, workers)
	for w := range sets {
		keys := make([]string, per)
		for i := range keys {
			keys[i] = fmt.Sprintf("w%d:key:%d", w, i)
		}
		sets[w] = keys
	}

	b.ResetTimer()

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(keys []string) {
			defer wg.Done()
			for i := 0; i < per; i++ {
				op(keys[i])
			}
		}(sets[w])
	}
	wg.Wait()

	b.StopTimer()
}

func BenchmarkShardedInsert(b *testing.B) {
	for _, shards := range []int{8, 64, 256} {
		for _, workers := range benchWorkerCounts {
			b.Run(fmt.Sprintf("shards=%d/workers=%d", shards, workers), func(b *testing.B) {
				s := newStoreUnpadded(shards)
				runInsertWorkers(b, workers, func(key string) { s.Set(key, "value") })
			})
		}
	}
}

func BenchmarkSyncMapInsert(b *testing.B) {
	for _, workers := range benchWorkerCounts {
		b.Run(fmt.Sprintf("workers=%d", workers), func(b *testing.B) {
			var m sync.Map
			runInsertWorkers(b, workers, func(key string) { m.Store(key, "value") })
		})
	}
}
