package wal

import (
	"errors"
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
