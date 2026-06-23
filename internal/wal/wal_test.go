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
