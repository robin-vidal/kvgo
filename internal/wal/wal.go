package wal

import (
	"bufio"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
)

type WAL struct {
	file   *os.File
	cfg    *config.WalConfig
	mu     sync.Mutex
	seqNum uint64
}

type Wal interface {
	Append(e Entry) error
	Close() error
	Replay() ([]Entry, error)
}

func Open(cfg *config.WalConfig) (*WAL, error) {
	file, err := os.OpenFile(cfg.WalPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &WAL{file: file, cfg: cfg}, nil
}

func (wal *WAL) Close() error {
	return wal.file.Close()
}

func (wal *WAL) Append(e Entry) error {
	if wal == nil || wal.file == nil {
		return errors.New("wal not initialized")
	}

	wal.mu.Lock()
	defer wal.mu.Unlock()

	wal.seqNum++
	e.SeqNum = wal.seqNum

	encoded, err := Encode(e)
	if err != nil {
		return err
	}

	n, err := wal.file.Write(encoded)
	if err != nil {
		return err
	}

	if n != len(encoded) {
		return errors.New("partial write to WAL")
	}

	return wal.file.Sync()
}

func (wal *WAL) Replay() ([]Entry, error) {
	file, err := os.Open(wal.cfg.WalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)
	var entries []Entry

	var offset int64 = 0

	for {
		entry, n, err := decodeOne(r)
		if err == io.EOF {
			break
		}

		if err != nil {
			if truncErr := os.Truncate(wal.cfg.WalPath, offset); truncErr != nil {
				return nil, truncErr
			}
			break
		}

		offset += int64(n)
		entries = append(entries, entry)
	}

	if len(entries) > 0 {
		wal.seqNum = entries[len(entries)-1].SeqNum
	}

	return entries, nil
}

func (wal *WAL) ShouldCompact() bool {
	return wal.seqNum%uint64(wal.cfg.CompactionThreshold) == 0
}
