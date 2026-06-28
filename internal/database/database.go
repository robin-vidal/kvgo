// Package database handles database initialization and data manipulation
package database

import (
	"hash/fnv"
	"maps"
	"sync"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
)

type databaseShard struct {
	mu   sync.RWMutex
	data map[string]string
}

// Database stores application data.
type Database struct {
	shards  []databaseShard
	metrics *metrics
}

// New creates and returns a new instance of the database.
func New(cfg *config.Config) (*Database, error) {
	db := &Database{
		shards: make([]databaseShard, cfg.ShardAmount),
	}

	for i := 0; i < cfg.ShardAmount; i++ {
		db.shards[i].data = make(map[string]string)
	}

	m, err := newMetrics()
	if err != nil {
		return nil, err
	}
	db.metrics = m

	return db, nil
}

// Set defines the value for a specific key in the map.
func (db *Database) Set(key, value string) {
	idx := getShard(key, len(db.shards))
	shard := &db.shards[idx]

	start := time.Now()
	shard.mu.Lock()
	waited := time.Since(start)
	shard.data[key] = value
	shard.mu.Unlock()

	db.metrics.record(idx, "set", "write", waited)
}

// Get retrieves the value in the map for a specific key.
func (db *Database) Get(key string) (string, bool) {
	idx := getShard(key, len(db.shards))
	shard := &db.shards[idx]

	start := time.Now()
	shard.mu.RLock()
	waited := time.Since(start)
	val, ok := shard.data[key]
	shard.mu.RUnlock()

	db.metrics.record(idx, "get", "read", waited)
	return val, ok
}

// Delete remove the key in the map.
func (db *Database) Delete(key string) {
	idx := getShard(key, len(db.shards))
	shard := &db.shards[idx]

	start := time.Now()
	shard.mu.Lock()
	waited := time.Since(start)
	delete(shard.data, key)
	shard.mu.Unlock()

	db.metrics.record(idx, "delete", "write", waited)
}

func getShard(key string, shardAmount int) int {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return int(hasher.Sum64() % uint64(shardAmount))
}

func (db *Database) GetKeyAmountPerShard() []int {
	amountPerShard := make([]int, 0, len(db.shards))

	for _, shard := range db.shards {
		shard.mu.RLock()
		amountPerShard = append(amountPerShard, len(shard.data))
		shard.mu.RUnlock()
	}

	return amountPerShard
}

func (db *Database) Dump() map[string]string {
	dump := make(map[string]string)

	for i := range db.shards {
		db.shards[i].mu.RLock()
		maps.Copy(dump, db.shards[i].data)
		db.shards[i].mu.RUnlock()
	}

	return dump
}
