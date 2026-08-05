package raft

import (
	"fmt"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/wal"
)

type Role int

const (
	Follower Role = iota
	Candidate
	Leader
)

type Node struct {
	state         *State
	cfg           *config.RaftConfig
	resetElection chan struct{}
	wal           *wal.WAL

	mu       sync.Mutex
	role     Role
	leaderID string
	votes    int
}

func NewNode(cfg *config.RaftConfig, wal *wal.WAL) (*Node, error) {
	state, err := LoadState(cfg)
	if err != nil {
		return nil, fmt.Errorf("raft: failed to load persistent state: %w", err)
	}

	return &Node{
		role:          Follower,
		resetElection: make(chan struct{}, 1),
		state:         state,
		cfg:           cfg,
		wal:           wal,
	}, nil
}
