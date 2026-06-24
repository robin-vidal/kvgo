package wal

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "simple open",
			path:    "test.wal",
			wantErr: false,
		},
		{
			name:    "unknown path",
			path:    "unknown/path.wal",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.path)
			cfg := &config.WalConfig{WalPath: path}

			wal, err := Open(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Open() error = %v, wanted = %v", err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if wal.cfg.WalPath != path {
				t.Errorf("Open() wal.path = %s, wanted = %s", wal.cfg.WalPath, path)
			}
		})
	}
}

func TestClose(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "simple close",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "wal.log")
			cfg := &config.WalConfig{WalPath: path}

			tmpWal, _ := Open(cfg)
			err := tmpWal.Close()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Close() error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}

func TestCloseTwice(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "close twice",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "wal.log")
			cfg := &config.WalConfig{WalPath: path}

			tmpWal, _ := Open(cfg)
			tmpWal.Close()
			err := tmpWal.Close()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Close() error = %v, want error = %v", err, tt.wantErr)
			}
		})
	}
}

func TestAppend(t *testing.T) {
	tests := []struct {
		name    string
		entries []Entry
		wantErr bool
	}{
		{
			name:    "single SET entry",
			entries: []Entry{{SeqNum: 1, Command: "SET", Key: "foo", Value: ptr("bar")}},
		},
		{
			name: "multiple entries accumulate",
			entries: []Entry{
				{SeqNum: 1, Command: "SET", Key: "foo", Value: ptr("bar")},
				{SeqNum: 2, Command: "DEL", Key: "foo"},
			},
		},
		{
			name:    "encode error propagates",
			entries: []Entry{{SeqNum: 1, Command: strings.Repeat("x", 256), Key: "foo", Value: ptr("bar")}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "wal.log")

			w, err := Open(&config.WalConfig{WalPath: path})
			if err != nil {
				t.Fatalf("Open() error = %v", err)
			}
			defer w.Close()

			var want []byte
			var lastErr error
			for _, e := range tt.entries {
				lastErr = w.Append(e)
				if lastErr != nil {
					break
				}
				encoded, _ := Encode(e)
				want = append(want, encoded...)
			}

			if (lastErr != nil) != tt.wantErr {
				t.Fatalf("Append() error = %v, wantErr %v", lastErr, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			got, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("ReadFile() error = %v", err)
			}

			if !bytes.Equal(got, want) {
				t.Errorf("file content mismatch:\ngot:  %v\nwant: %v", got, want)
			}
		})
	}
}

func TestReplay_NoExistingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.log")
	w := &WAL{cfg: &config.WalConfig{WalPath: path}}

	entries, err := w.Replay()
	if err != nil || entries != nil {
		t.Fatalf("Replay() = %v, %v, want nil, nil", entries, err)
	}
}

func TestReplay_AllValidEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	w, _ := Open(&config.WalConfig{WalPath: path})

	want := []Entry{
		{Command: "SET", Key: "foo", Value: ptr("bar")},
		{Command: "DEL", Key: "foo"},
		{Command: "SET", Key: "baz", Value: ptr("qux")},
	}
	for _, e := range want {
		w.Append(e)
	}
	w.Close()

	w2, _ := Open(&config.WalConfig{WalPath: path})
	defer w2.Close()

	got, err := w2.Replay()
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("Replay() = %d entries, want %d", len(got), len(want))
	}
	for i, e := range got {
		if e.Command != want[i].Command || e.Key != want[i].Key || !valueEqual(e.Value, want[i].Value) {
			t.Errorf("entry %d = %+v, want %+v", i, e, want[i])
		}
	}
}

func TestReplay_TruncatesCorruptedTrailingEntry(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	w, _ := Open(&config.WalConfig{WalPath: path})
	w.Append(Entry{Command: "SET", Key: "foo", Value: ptr("bar")})
	w.Close()

	validSize, _ := os.Stat(path)

	f, _ := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0644)
	f.Write([]byte{0, 0, 0, 0, 0, 0, 0, 2, 3})
	f.Close()

	w2, _ := Open(&config.WalConfig{WalPath: path})
	defer w2.Close()

	got, err := w2.Replay()
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(got) != 1 || got[0].Key != "foo" {
		t.Fatalf("Replay() = %+v, want 1 entry with Key=foo", got)
	}

	info, _ := os.Stat(path)
	if info.Size() != validSize.Size() {
		t.Errorf("size = %d, want %d", info.Size(), validSize.Size())
	}
}

func TestReplay_SeqNumContinuesAfterRestart(t *testing.T) {
	path := filepath.Join(t.TempDir(), "wal.log")
	w, _ := Open(&config.WalConfig{WalPath: path})
	w.Append(Entry{Command: "SET", Key: "a", Value: ptr("1")})
	w.Append(Entry{Command: "SET", Key: "b", Value: ptr("2")})
	w.Close()

	w2, _ := Open(&config.WalConfig{WalPath: path})
	if _, err := w2.Replay(); err != nil {
		t.Fatalf("Replay() error = %v", err)
	}

	if err := w2.Append(Entry{Command: "SET", Key: "c", Value: ptr("3")}); err != nil {
		t.Fatalf("Append() error = %v", err)
	}
	w2.Close()

	w3, _ := Open(&config.WalConfig{WalPath: path})
	defer w3.Close()

	entries, err := w3.Replay()
	if err != nil {
		t.Fatalf("Replay() error = %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("got %d entries, want 3", len(entries))
	}
	if entries[2].SeqNum != 3 {
		t.Errorf("third entry SeqNum = %d, want 3", entries[2].SeqNum)
	}
}
