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

	mu       sync.RWMutex
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

func (n *Node) CurrentTerm() uint64 {
	n.state.mu.RLock()
	defer n.state.mu.RUnlock()

	return n.state.CurrentTerm
}

func (n *Node) State() Role {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.role
}

func (n *Node) becomeFollower(term uint64) {
	n.mu.Lock()
	n.role = Follower
	n.mu.Unlock()

	newTerm := max(term, n.CurrentTerm())
	n.state.SetTermAndVote(newTerm, "")
}

func (n *Node) becomeCandidate() {
	n.state.SetTermAndVote(n.CurrentTerm()+1, n.cfg.NodeID)

	n.mu.Lock()
	defer n.mu.Unlock()
	n.role = Candidate
	n.votes = 1
}

func (n *Node) becomeLeader() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.role = Leader
	n.leaderID = n.cfg.NodeID
}
