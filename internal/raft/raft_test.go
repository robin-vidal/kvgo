package raft

import (
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
)

func TestNewNode(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "valid path",
			path:    "raft.state",
			wantErr: false,
		},
		{
			name:    "creates missing parent directory",
			path:    "subdir/raft.state",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.path)
			cfg := &config.RaftConfig{RaftStatePath: path}

			node, err := NewNode(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("NewNode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && node == nil {
				t.Error("NewNode() returned nil node")
			}
			if node.role != Follower {
				t.Errorf("role = %v, want Follower", node.role)
			}
			if cap(node.resetElection) != 1 {
				t.Errorf("resetElection cap = %d, want 1", cap(node.resetElection))
			}
		})
	}
}
