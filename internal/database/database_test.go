package database

import (
	"fmt"
	"sync"
	"testing"
)

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
			db := New()
			db.data = tt.data

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
			db := New()
			db.data = tt.data

			db.Set(tt.key, tt.wantVal)

			val, ok := tt.data[tt.key]
			if val != tt.wantVal {
				t.Errorf("Get() val = %v, wantVal %v", val, tt.wantVal)
			}
			if ok != true {
				t.Errorf("Get() ok = %v, want %v", ok, true)
			}
		})
	}
}

func TestDatabaseConcurrency(t *testing.T) {
	db := New()
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
