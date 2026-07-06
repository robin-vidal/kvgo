package wal

import (
	"bufio"
	"errors"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
)

type WAL struct {
	file    *os.File
	cfg     *config.WalConfig
	mu      sync.Mutex
	seqNum  uint64
	metrics *metrics
}

type Wal interface {
	Append(e Entry) error
	Close() error
	Replay() ([]Entry, error)
}

func Open(cfg *config.WalConfig) (*WAL, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.WalPath), 0o755); err != nil {
		return nil, err
	}

	file, err := os.OpenFile(cfg.WalPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	wal := &WAL{file: file, cfg: cfg}

	m, err := newMetrics(wal)
	if err != nil {
		return nil, err
	}
	wal.metrics = m

	return wal, nil
}

func (wal *WAL) Close() error {
	return wal.file.Close()
}

func (wal *WAL) Append(e Entry) (err error) {
	if wal == nil || wal.file == nil {
		return errors.New("wal not initialized")
	}

	start := time.Now()
	defer func() {
		wal.metrics.recordAppend(time.Since(start), err)
	}()

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

func (wal *WAL) Replay() (entries []Entry, err error) {
	start := time.Now()
	defer func() {
		wal.metrics.recordReplay(time.Since(start), len(entries))
	}()

	file, err := os.Open(wal.cfg.WalPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	r := bufio.NewReader(file)

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
	if wal.cfg.CompactionThreshold <= 0 {
		return false
	}

	wal.mu.Lock()
	defer wal.mu.Unlock()

	return wal.seqNum > 0 && wal.seqNum%uint64(wal.cfg.CompactionThreshold) == 0
}

func (wal *WAL) CurrentSeqNum() uint64 {
	wal.mu.Lock()
	defer wal.mu.Unlock()
	return wal.seqNum
}
