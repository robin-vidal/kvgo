package raft

import (
	"path/filepath"
	"sync"
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
			path:    "missing/raft.state",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), tt.path)
			cfg := &config.RaftConfig{RaftStatePath: path}

			node, err := NewNode(cfg, nil)
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

func TestState(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*Node)
		want       Role
	}{
		{
			name:       "become follower",
			transition: func(n *Node) { n.becomeFollower(0) },
			want:       Follower,
		},
		{
			name:       "become candidate",
			transition: func(n *Node) { n.becomeCandidate() },
			want:       Candidate,
		},
		{
			name:       "become leader",
			transition: func(n *Node) { n.becomeLeader() },
			want:       Leader,
		},
		{
			name: "candidate to follower",
			transition: func(n *Node) {
				n.becomeCandidate()
				n.becomeFollower(1)
			},
			want: Follower,
		},
		{
			name: "leader to follower",
			transition: func(n *Node) {
				n.becomeLeader()
				n.becomeFollower(1)
			},
			want: Follower,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "raft.state")
			cfg := &config.RaftConfig{RaftStatePath: path}

			node, err := NewNode(cfg, nil)
			if err != nil {
				t.Fatalf("NewNode() error = %v", err)
			}

			tt.transition(node)

			if tt.want != node.State() {
				t.Errorf("role = %v, want %v", node.role, tt.want)
			}
		})
	}
}

func TestCurrentTerm(t *testing.T) {
	tests := []struct {
		name       string
		transition func(*Node)
		want       uint64
	}{
		{
			name:       "initial term is zero",
			transition: func(n *Node) {},
			want:       0,
		},
		{
			name:       "becomeCandidate increments term",
			transition: func(n *Node) { n.becomeCandidate() },
			want:       1,
		},
		{
			name: "becomeFollower with higher term updates term",
			transition: func(n *Node) {
				n.becomeCandidate()
				n.becomeFollower(5)
			},
			want: 5,
		},
		{
			name: "becomeFollower with lower term keeps current",
			transition: func(n *Node) {
				n.becomeCandidate()
				n.becomeFollower(0)
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "raft.state")
			cfg := &config.RaftConfig{RaftStatePath: path}

			node, err := NewNode(cfg, nil)
			if err != nil {
				t.Fatalf("NewNode() error = %v", err)
			}

			tt.transition(node)

			if got := node.CurrentTerm(); got != tt.want {
				t.Errorf("CurrentTerm() = %v, want %v", got, tt.want)
			}
		})
	}
}

func newTestNode(t *testing.T) *Node {
	t.Helper()
	path := filepath.Join(t.TempDir(), "raft.state")
	cfg := &config.RaftConfig{RaftStatePath: path}
	node, err := NewNode(cfg, nil)
	if err != nil {
		t.Fatalf("NewNode() error = %v", err)
	}
	return node
}

func TestStateConcurrent(t *testing.T) {
	node := newTestNode(t)

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.State()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.becomeCandidate()
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.becomeLeader()
	}()
	wg.Wait()
}

func TestCurrentTermConcurrent(t *testing.T) {
	node := newTestNode(t)

	var wg sync.WaitGroup
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.CurrentTerm()
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.becomeCandidate()
	}()
	wg.Wait()
}
