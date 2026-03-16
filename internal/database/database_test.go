package database

import (
	"fmt"
	"sync"
	"testing"

	"github.com/rvHoney/kvgo/internal/config"
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
