// Package database handles database initialization and data manipulation
package database

import "sync"

// Database stores application data and avoids race conditions.
type Database struct {
	mu   sync.RWMutex
	data map[string]string
}

// New creates and returns a new instance of the database.
func New() *Database {
	return &Database{
		data: make(map[string]string),
	}
}

// Set defines the value for a specific key in the map.
func (db *Database) Set(key, value string) {
	db.mu.Lock()
	defer db.mu.Unlock()
	db.data[key] = value
}

// Get retrieves the value in the map for a specific key.
func (db *Database) Get(key string) (string, bool) {
	db.mu.RLock()
	defer db.mu.RUnlock()
	val, ok := db.data[key]
	return val, ok
}
