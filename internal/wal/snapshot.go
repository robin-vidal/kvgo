package wal

import (
	"bufio"
	"encoding/gob"
	"io"
	"os"
	"path/filepath"
)

func (wal *WAL) Snapshot(data map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(wal.cfg.SnapshotPath), 0o755); err != nil {
		return err
	}

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

func (wal *WAL) truncate(snapshotSeqNum uint64) error {
	// Read WAL
	file, err := os.Open(wal.cfg.WalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	r := bufio.NewReader(file)
	var entries []Entry
	for {
		entry, _, err := decodeOne(r)
		if err == io.EOF {
			break
		}
		if err != nil {
			file.Close()
			return err
		}
		entries = append(entries, entry)
	}
	file.Close()

	// Create temporary WAL
	tmpPath := wal.cfg.WalPath + ".tmp"
	tmpFile, err := os.OpenFile(tmpPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.SeqNum <= snapshotSeqNum {
			continue
		}
		encoded, err := Encode(e)
		if err != nil {
			tmpFile.Close()
			return err
		}
		if _, err := tmpFile.Write(encoded); err != nil {
			tmpFile.Close()
			return err
		}
	}

	if err := tmpFile.Sync(); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	// Replace WAL by temporary WAL
	wal.mu.Lock()
	defer wal.mu.Unlock()

	if err := wal.file.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpPath, wal.cfg.WalPath); err != nil {
		return err
	}

	newFile, err := os.OpenFile(wal.cfg.WalPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	wal.file = newFile

	return nil
}

func (wal *WAL) Compact(data map[string]string, snapshotSeqNum uint64) error {
	if err := wal.Snapshot(data); err != nil {
		return err
	}

	return wal.truncate(snapshotSeqNum)
}

func (wal *WAL) LoadSnapshot() (map[string]string, error) {
	f, err := os.Open(wal.cfg.SnapshotPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer f.Close()

	var data map[string]string
	if err := gob.NewDecoder(f).Decode(&data); err != nil {
		return nil, err
	}
	return data, nil
}
