//go:build !windows

// Package fsutil provides small filesystem helpers with platform-specific
// behavior.
package fsutil

import "os"

// SyncDir fsyncs the directory so a preceding rename is durably persisted.
func SyncDir(path string) error {
	dir, err := os.Open(path)
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
