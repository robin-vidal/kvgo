package raft

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/robin-vidal/kvgo/internal/config"
)

func TestLoadState(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(path string)
		wantTerm uint64
		wantVote string
		wantErr  bool
	}{
		{
			name:     "fresh state",
			wantTerm: 0,
			wantVote: "",
			wantErr:  false,
		},
		{
			name: "loads persisted state",
			setup: func(path string) {
				os.WriteFile(path, []byte(`{"current_term":5,"voted_for":"node-1"}`), 0o600)
			},
			wantTerm: 5,
			wantVote: "node-1",
			wantErr:  false,
		},
		{
			name: "corrupted json",
			setup: func(path string) {
				os.WriteFile(path, []byte(`not-json`), 0o600)
			},
			wantErr: true,
		},
		{
			name:     "creates missing parent directory",
			wantTerm: 0,
			wantVote: "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			var path string
			if tt.name == "creates missing parent directory" {
				path = filepath.Join(dir, "subdir", "raft.state")
			} else {
				path = filepath.Join(dir, "raft.state")
			}

			if tt.setup != nil {
				tt.setup(path)
			}

			cfg := &config.RaftConfig{RaftStatePath: path}
			s, err := LoadState(cfg)
			if (err != nil) != tt.wantErr {
				t.Fatalf("LoadState() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}

			if s.CurrentTerm != tt.wantTerm {
				t.Errorf("LoadState() CurrentTerm = %d, want %d", s.CurrentTerm, tt.wantTerm)
			}
			if s.VotedFor != tt.wantVote {
				t.Errorf("LoadState() VotedFor = %q, want %q", s.VotedFor, tt.wantVote)
			}

			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("LoadState() did not create state file at %s", path)
			}
		})
	}
}

func TestSetTermAndVote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raft.state")
	cfg := &config.RaftConfig{RaftStatePath: path}

	s, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if err := s.SetTermAndVote(3, "node-2"); err != nil {
		t.Fatalf("SetTermAndVote() error = %v", err)
	}

	s2, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() after SetTermAndVote error = %v", err)
	}

	if s2.CurrentTerm != 3 {
		t.Errorf("CurrentTerm = %d, want 3", s2.CurrentTerm)
	}
	if s2.VotedFor != "node-2" {
		t.Errorf("VotedFor = %q, want %q", s2.VotedFor, "node-2")
	}
}

func TestSetTerm(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raft.state")
	cfg := &config.RaftConfig{RaftStatePath: path}

	s, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if err := s.SetTermAndVote(1, "node-1"); err != nil {
		t.Fatalf("SetTermAndVote() error = %v", err)
	}
	if err := s.SetTerm(7); err != nil {
		t.Fatalf("SetTerm() error = %v", err)
	}

	s2, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() after SetTerm error = %v", err)
	}

	if s2.CurrentTerm != 7 {
		t.Errorf("CurrentTerm = %d, want 7", s2.CurrentTerm)
	}
	if s2.VotedFor != "node-1" {
		t.Errorf("VotedFor = %q, want %q (vote must be preserved)", s2.VotedFor, "node-1")
	}
}

func TestSetVote(t *testing.T) {
	path := filepath.Join(t.TempDir(), "raft.state")
	cfg := &config.RaftConfig{RaftStatePath: path}

	s, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() error = %v", err)
	}

	if err := s.SetTermAndVote(2, "node-1"); err != nil {
		t.Fatalf("SetTermAndVote() error = %v", err)
	}
	if err := s.SetVote("node-3"); err != nil {
		t.Fatalf("SetVote() error = %v", err)
	}

	s2, err := LoadState(cfg)
	if err != nil {
		t.Fatalf("LoadState() after SetVote error = %v", err)
	}

	if s2.VotedFor != "node-3" {
		t.Errorf("VotedFor = %q, want %q", s2.VotedFor, "node-3")
	}
	if s2.CurrentTerm != 2 {
		t.Errorf("CurrentTerm = %d, want 2 (term must be preserved)", s2.CurrentTerm)
	}
}
