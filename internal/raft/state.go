package raft

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
)

type State struct {
	mu   sync.RWMutex
	path string

	CurrentTerm uint64 `json:"current_term"`
	VotedFor    string `json:"voted_for"`
}

func LoadState(cfg *config.RaftConfig) (*State, error) {
	if err := os.MkdirAll(filepath.Dir(cfg.RaftStatePath), 0o755); err != nil {
		return nil, err
	}

	s := &State{path: cfg.RaftStatePath}

	data, err := os.ReadFile(cfg.RaftStatePath)
	switch {
	case os.IsNotExist(err):
		if err := s.persist(); err != nil {
			return nil, err
		}
		return s, nil
	case err != nil:
		return nil, err
	}

	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *State) persist() error {
	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	tmpPath := s.path + ".tmp"
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, s.path); err != nil {
		return err
	}

	dir, err := os.Open(filepath.Dir(s.path))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}

func (s *State) SetTermAndVote(term uint64, votedFor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.CurrentTerm = term
	s.VotedFor = votedFor
	return s.persist()
}

func (s *State) SetTerm(term uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.CurrentTerm = term
	return s.persist()
}

func (s *State) SetVote(votedFor string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.VotedFor = votedFor
	return s.persist()
}
