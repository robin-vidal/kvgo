package database

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
)

func generateSampleConfig(shardAmount int) *config.Config {
	return &config.Config{
		Host:        "localhost",
		Port:        6379,
		Debug:       false,
		ShardAmount: shardAmount,
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		key     string
		wantVal string
		wantOk  bool
	}{
		{
			name:    "Existing Key",
			data:    map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:     "field1",
			wantVal: "value1",
			wantOk:  true,
		},
		{
			name:    "Not Found",
			data:    map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:     "field4'",
			wantVal: "",
			wantOk:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := generateSampleConfig(2)
			db := New(cfg)

			for key, val := range tt.data {
				db.Set(key, val)
			}

			val, ok := db.Get(tt.key)
			if val != tt.wantVal {
				t.Errorf("Get() val = %v, wantVal %v", val, tt.wantVal)
			}
			if ok != tt.wantOk {
				t.Errorf("Get() ok = %v, wantOk %v", ok, tt.wantOk)
			}
		})
	}
}

func TestSet(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]string
		key     string
		wantVal string
	}{
		{
			name:    "Already Exist",
			data:    map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:     "field1",
			wantVal: "value1",
		},
		{
			name:    "New Key",
			data:    map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:     "field4",
			wantVal: "field4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := generateSampleConfig(2)
			db := New(cfg)

			for key, val := range tt.data {
				db.Set(key, val)
			}

			db.Set(tt.key, tt.wantVal)

			val, ok := db.Get(tt.key)
			if val != tt.wantVal {
				t.Errorf("Get() val = %v, wantVal %v", val, tt.wantVal)
			}
			if !ok {
				t.Errorf("Get() ok = %v, want %v", ok, true)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	tests := []struct {
		name string
		data map[string]string
		key  string
	}{
		{
			name: "Existing Key",
			data: map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:  "field2",
		},
		{
			name: "Unknow Key",
			data: map[string]string{"field1": "value1", "field2": "value2", "field3": "value3"},
			key:  "field4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := generateSampleConfig(2)
			db := New(cfg)

			for key, val := range tt.data {
				db.Set(key, val)
			}

			db.Delete(tt.key)

			val, ok := db.Get(tt.key)
			if ok {
				t.Errorf("Get() val = %v, want <nil>", val)
			}
		})
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	cfg := generateSampleConfig(2)
	db := New(cfg)

	const workers int = 50
	const iterations int = 1000

	var wg sync.WaitGroup
	wg.Add(workers * 2)

	// write test
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d", j%10)
				db.Set(key, "value")
			}
		}(i)
	}

	// read test
	for i := 0; i < workers; i++ {
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				key := fmt.Sprintf("key-%d", j%10)
				db.Get(key)
			}
		}(i)
	}

	wg.Wait()
}

func TestGetShard(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		shardAmount int
		wantVal     int
	}{
		{
			name:        "Basic key 1",
			key:         "salut",
			shardAmount: 2,
			wantVal:     0,
		},
		{
			name:        "Basic key 2",
			key:         "test",
			shardAmount: 2,
			wantVal:     1,
		},
		{
			name:        "Huge Shard Amount",
			key:         "BIIIG",
			shardAmount: 200,
			wantVal:     95,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			val := getShard(tt.key, tt.shardAmount)
			if val != tt.wantVal {
				t.Errorf("getShard() val = %v, want %v", val, tt.wantVal)
			}
		})
	}
}

func TestGetKeyAmountPerShard(t *testing.T) {
	tests := []struct {
		name        string
		shardAmount int
		data        map[string]string
		wantTotal   int
	}{
		{
			name:        "Empty Database",
			shardAmount: 4,
			data:        make(map[string]string),
			wantTotal:   0,
		},
		{
			name:        "Simple Database",
			shardAmount: 4,
			data:        map[string]string{"a": "1", "b": "2", "c": "3"},
			wantTotal:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := generateSampleConfig(tt.shardAmount)
			db := New(cfg)

			for key, value := range tt.data {
				db.Set(key, value)
			}

			amountPerShard := db.GetKeyAmountPerShard()
			totalAmount := 0
			for _, currentAmount := range amountPerShard {
				totalAmount += currentAmount
			}

			if totalAmount != tt.wantTotal {
				t.Errorf("GetKeyAmountPerShard() val = %v, want %v", totalAmount, tt.wantTotal)
			}

			if len(amountPerShard) != tt.shardAmount {
				t.Errorf("GetKeyAmountPerShard() len() = %v, want %v", len(amountPerShard), tt.shardAmount)
			}
		})
	}
}

// TestGetKeyAmountPerShard_RaceWithSetDel guards issue #28: prior to the
// pointer-iter fix, GetKeyAmountPerShard ranged over db.shards by value,
// copying each shard (including its sync.RWMutex) so the RLock was taken
// on the per-iteration copy and the real shard stayed unlocked. That
// reliably fired -race against concurrent Set/Del.
func TestGetKeyAmountPerShard_RaceWithSetDel(t *testing.T) {
	db := New(generateSampleConfig(8))

	var wg sync.WaitGroup
	stop := make(chan struct{})

	wg.Add(1)
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-stop:
				return
			default:
			}
			k := fmt.Sprintf("k%d", i%128)
			db.Set(k, "v")
			db.Delete(k)
			i++
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for j := 0; j < 2000; j++ {
			select {
			case <-stop:
				return
			default:
			}
			_ = db.GetKeyAmountPerShard()
		}
	}()

	// Let the workload run briefly then signal stop. The race detector
	// fires at goroutine-exit time if any interleaving occurred, so the
	// duration just needs to be long enough to produce overlap.
	time.Sleep(50 * time.Millisecond)
	close(stop)
	wg.Wait()
}
