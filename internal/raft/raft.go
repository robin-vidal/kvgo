package raft

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/raft/raftpb"
	"github.com/robin-vidal/kvgo/internal/wal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	grpcServer    *grpc.Server
	peers         map[string]raftpb.RaftServiceClient
	conns         []*grpc.ClientConn

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

func (n *Node) Start() error {
	address := fmt.Sprintf("%s:%d", n.cfg.Host, n.cfg.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}

	slog.Info("raft gRPC server is listening", "addr", ln.Addr().String())

	n.grpcServer = grpc.NewServer()
	raftpb.RegisterRaftServiceServer(n.grpcServer, &grpcTransport{node: n})
	go n.grpcServer.Serve(ln)

	return n.dialPeers()
}

func (n *Node) Stop() {
	if n.grpcServer != nil {
		n.grpcServer.GracefulStop()
	}

	for _, conn := range n.conns {
		conn.Close()
	}
}

func (n *Node) dialPeers() error {
	n.peers = make(map[string]raftpb.RaftServiceClient)
	for _, peerAddr := range n.cfg.Peers {
		conn, err := grpc.NewClient(peerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			return err
		}

		n.conns = append(n.conns, conn)
		n.peers[peerAddr] = raftpb.NewRaftServiceClient(conn)
	}

	return nil
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

func (n *Node) sendRequestVote(peer string, req *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {

	p, found := n.peers[peer]
	if !found {
		return nil, fmt.Errorf("raft: peer %q not found", peer)
	}

	ctx, cancel := context.WithTimeout(context.Background(), n.cfg.RPCTimeout)
	defer cancel()

	resp, err := p.RequestVote(ctx, req)
	return resp, err
}

func (n *Node) sendAppendEntries(peer string, req *raftpb.AppendEntriesRequest) (*raftpb.AppendEntriesResponse, error) {
	return nil, nil
}
