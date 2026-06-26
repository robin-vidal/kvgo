package wal

import (
	"encoding/gob"
	"os"
)

func (wal *WAL) Snapshot(data map[string]string) error {
	tmpPath := wal.cfg.SnapshotPath + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	if err := gob.NewEncoder(f).Encode(data); err != nil {
		f.Close()
		return err
	}

	if err := f.Sync(); err != nil {
		f.Close()
		return err
	}

	if err := f.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, wal.cfg.SnapshotPath)
}

func (wal *WAL) truncate() error {
	return nil
}

func (wal *WAL) Compact(data map[string]string) error {
	if err := wal.Snapshot(data); err != nil {
		return err
	}

	return wal.truncate()
}
