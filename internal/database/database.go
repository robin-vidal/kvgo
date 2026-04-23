// Package database handles database initialization and data manipulation
package database

import (
	"hash/fnv"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
)

type databaseShard struct {
	mu   sync.RWMutex
	data map[string]string
}

// Database stores application data.
type Database struct {
	shards []databaseShard
}

// New creates and returns a new instance of the database.
func New(cfg *config.Config) *Database {
	db := &Database{
		shards: make([]databaseShard, cfg.ShardAmount),
	}

	for i := 0; i < cfg.ShardAmount; i++ {
		db.shards[i].data = make(map[string]string)
	}

	return db
}

// Set defines the value for a specific key in the map.
func (db *Database) Set(key, value string) {
	shard := &db.shards[getShard(key, len(db.shards))]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	shard.data[key] = value
}

// Get retrieves the value in the map for a specific key.
func (db *Database) Get(key string) (string, bool) {
	shard := &db.shards[getShard(key, len(db.shards))]

	shard.mu.RLock()
	defer shard.mu.RUnlock()
	val, ok := shard.data[key]
	return val, ok
}

// Delete remove the key in the map.
func (db *Database) Delete(key string) {
	shard := &db.shards[getShard(key, len(db.shards))]

	shard.mu.Lock()
	defer shard.mu.Unlock()
	delete(shard.data, key)
}

func getShard(key string, shardAmount int) int {
	hasher := fnv.New64a()
	hasher.Write([]byte(key))
	return int(hasher.Sum64() % uint64(shardAmount))
}

func (db *Database) GetKeyAmountPerShard() []int {
	amountPerShard := make([]int, 0, len(db.shards))

	// Iterate by index so shard is a pointer to the real slice element.
	// The previous `for _, shard := range db.shards` copied the
	// databaseShard value (including its sync.RWMutex), so the RLock was
	// taken on a per-iteration copy and left the real shard unlocked;
	// concurrent Set/Del calls hit by -race and could see a torn
	// len(shard.data).
	for i := range db.shards {
		shard := &db.shards[i]
		shard.mu.RLock()
		amountPerShard = append(amountPerShard, len(shard.data))
		shard.mu.RUnlock()
	}

	return amountPerShard
}
