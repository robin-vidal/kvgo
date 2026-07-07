package raft

import (
	"sync"
)

type State struct {
	mu   sync.Mutex
	path string

	CurrentTerm uint64 `json:"current_term"`
	VotedFor    string `json:"voted_for"`
}
