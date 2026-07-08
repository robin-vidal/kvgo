package raft

import (
	"fmt"

	"github.com/robin-vidal/kvgo/internal/config"
)

type Node struct {
	state *State
	cfg   *config.RaftConfig
}

func NewNode(cfg *config.RaftConfig) (*Node, error) {
	state, err := LoadState(cfg)
	if err != nil {
		return nil, fmt.Errorf("raft: failed to load persistent state: %w", err)
	}

	return &Node{
		state: state,
		cfg:   cfg,
	}, nil
}
