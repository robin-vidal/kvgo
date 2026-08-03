//go:build windows

// Package fsutil provides small filesystem helpers with platform-specific
// behavior.
package fsutil

// SyncDir is a no-op on Windows: FlushFileBuffers on a directory handle
// returns ERROR_ACCESS_DENIED, and NTFS has no directory-fsync equivalent.
func SyncDir(path string) error {
	return nil
}
