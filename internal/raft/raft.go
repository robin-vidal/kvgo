package raft

import (
	"fmt"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
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

	mu       sync.Mutex
	role     Role
	leaderID string
	votes    int
}

func NewNode(cfg *config.RaftConfig) (*Node, error) {
	state, err := LoadState(cfg)
	if err != nil {
		return nil, fmt.Errorf("raft: failed to load persistent state: %w", err)
	}

	return &Node{
		role:          Follower,
		resetElection: make(chan struct{}, 1),
		state:         state,
		cfg:           cfg,
	}, nil
}
