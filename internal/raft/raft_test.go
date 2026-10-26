package raft

import (
	"context"
	"net"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/raft/raftpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

func TestElectionTimeout_ExpiresBecomesCandidate(t *testing.T) {
	n, _ := newStartedTestNode(t)
	time.Sleep(n.cfg.ElectionTimeoutMax + 50*time.Millisecond)

	if n.State() != Candidate {
		t.Fatalf("State() == %v, want Candidate", n.State())
	}
}

func TestElectionTimeout_ResetOnHeartbeat(t *testing.T) {
	n, addr := newStartedTestNode(t)
	time.Sleep(n.cfg.ElectionTimeoutMin - 50*time.Millisecond)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	client := raftpb.NewRaftServiceClient(conn)
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				client.AppendEntries(context.Background(), &raftpb.AppendEntriesRequest{
					Term: 1, LeaderId: "leader-1",
				})
			case <-stop:
				return
			}
		}
	}()

	time.Sleep(500 * time.Millisecond)
	close(stop)

	if n.State() != Follower {
		t.Fatalf("State() == %v, want Follower", n.State())
	}
}

func TestHeartbeat_leaderSendsToPeers(t *testing.T) {
	n1, _ := newStartedTestNode(t)
	n2, addr2 := newStartedTestNode(t)

	conn, err := grpc.NewClient(addr2, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	n1.peers[addr2] = raftpb.NewRaftServiceClient(conn)
	n1.becomeLeader()

	time.Sleep(n1.cfg.HeartbeatInterval * 3)

	n2.mu.RLock()
	leaderID := n2.leaderID
	n2.mu.RUnlock()

	if leaderID != n1.cfg.NodeID {
		t.Fatalf("leaderID == %v, want %v", leaderID, n1.cfg.NodeID)
	}
}

type voteGrantedService struct {
	raftpb.UnimplementedRaftServiceServer
}

func (s *voteGrantedService) RequestVote(_ context.Context, _ *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {
	return &raftpb.RequestVoteResponse{VoteGranted: true, Term: 1}, nil
}

func startFakeVoter(t *testing.T, v raftpb.RaftServiceServer) string {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	srv := grpc.NewServer()

	raftpb.RegisterRaftServiceServer(srv, v)
	go srv.Serve(ln)
	t.Cleanup(srv.GracefulStop)

	return ln.Addr().String()
}

func TestElection_MajorityBecomesLeader(t *testing.T) {
	peer1Addr := startFakeVoter(t, &voteGrantedService{})
	peer2Addr := startFakeVoter(t, &voteGrantedService{})

	n, _ := newUnstartedTestNode(t)
	n.cfg.Peers = []string{peer1Addr, peer2Addr}

	if err := n.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { n.Stop() })

	time.Sleep(n.cfg.ElectionTimeoutMax + 50*time.Millisecond)

	if n.State() != Leader {
		t.Fatalf("State() == %v, want Leader", n.State())
	}
}

type voteDeniedService struct {
	raftpb.UnimplementedRaftServiceServer
}

func (s *voteDeniedService) RequestVote(_ context.Context, _ *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {
	return &raftpb.RequestVoteResponse{VoteGranted: false, Term: 1}, nil
}

func TestElection_NoMajorityStaysCandidate(t *testing.T) {
	peer1Addr := startFakeVoter(t, &voteDeniedService{})
	peer2Addr := startFakeVoter(t, &voteDeniedService{})

	n, _ := newUnstartedTestNode(t)
	n.cfg.Peers = []string{peer1Addr, peer2Addr}

	if err := n.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { n.Stop() })

	time.Sleep(n.cfg.ElectionTimeoutMax + 50*time.Millisecond)

	if n.State() != Candidate {
		t.Fatalf("State() == %v, want Candidate", n.State())
	}
}

func TestElection_AppendEntriesStopsElection(t *testing.T) {
	peer1Addr := startFakeVoter(t, &slowRaftService{})
	peer2Addr := startFakeVoter(t, &slowRaftService{})

	n, addr := newUnstartedTestNode(t)
	n.cfg.Peers = []string{peer1Addr, peer2Addr}

	if err := n.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { n.Stop() })

	time.Sleep(n.cfg.ElectionTimeoutMax + 50*time.Millisecond)

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	client := raftpb.NewRaftServiceClient(conn)
	client.AppendEntries(context.Background(), &raftpb.AppendEntriesRequest{
		Term:     n.CurrentTerm(),
		LeaderId: "leader-id",
	})

	time.Sleep(50 * time.Millisecond)

	if n.State() != Follower {
		t.Fatalf("State() == %v, want Follower", n.State())
	}
}
