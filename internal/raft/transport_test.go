package raft

import (
	"context"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/raft/raftpb"
	"github.com/robin-vidal/kvgo/internal/wal"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

func newUnstartedTestNode(t *testing.T) (*Node, string) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	port := l.Addr().(*net.TCPAddr).Port
	l.Close()

	walCfg := config.WalConfig{
		WalPath:             t.TempDir() + "/wal.log",
		SnapshotPath:        t.TempDir() + "/snapshot.db",
		CompactionThreshold: 100, // anything should work
	}

	wal, err := wal.Open(&walCfg)
	if err != nil {
		t.Fatal(err)
	}

	raftCfg := config.RaftConfig{
		NodeID:             "node-" + strconv.Itoa(rand.Int()),
		Host:               "127.0.0.1",
		Port:               port,
		RPCTimeout:         10 * time.Millisecond,
		ElectionTimeoutMin: 100 * time.Millisecond,
		ElectionTimeoutMax: 200 * time.Millisecond,
		HeartbeatInterval:  100 * time.Millisecond,
		RaftStatePath:      t.TempDir() + "/raft.state",
	}
	addr := raftCfg.Host + ":" + strconv.Itoa(raftCfg.Port)

	node, err := NewNode(&raftCfg, wal)
	if err != nil {
		t.Fatal(err)
	}

	return node, addr
}

func newStartedTestNode(t *testing.T) (*Node, string) {
	node, addr := newUnstartedTestNode(t)
	err := node.Start()
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() { node.Stop() })

	return node, addr
}

type slowRaftService struct {
	raftpb.UnimplementedRaftServiceServer
}

func (s *slowRaftService) RequestVote(ctx context.Context, req *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {
	select {
	case <-time.After(10 * time.Second):
		return &raftpb.RequestVoteResponse{}, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func TestRPC_RequestVote(t *testing.T) {
	node, addr := newStartedTestNode(t)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	client := raftpb.NewRaftServiceClient(conn)

	req := raftpb.RequestVoteRequest{
		Term:        1,
		CandidateId: node.cfg.NodeID,
	}
	resp, err := client.RequestVote(context.Background(), &req)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.VoteGranted {
		t.Errorf("VoteGranted = false, want true")
	}
	if resp.Term != 1 {
		t.Errorf("Term = %d, want 1", resp.Term)
	}
}

func TestRPC_AppendEntries(t *testing.T) {
	_, addr := newStartedTestNode(t)
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	client := raftpb.NewRaftServiceClient(conn)

	req := raftpb.AppendEntriesRequest{
		Term:     1,
		LeaderId: "leader-id",
	}
	resp, err := client.AppendEntries(context.Background(), &req)
	if err != nil {
		t.Fatal(err)
	}

	if !resp.Success {
		t.Errorf("Success = false, want true")
	}
	if resp.Term != 1 {
		t.Errorf("Term = %d, want 1", resp.Term)
	}
}

func TestRPC_UnreachablePeer(t *testing.T) {
	node, _ := newUnstartedTestNode(t)
	deadAddr := "127.0.0.1:1"
	conn, err := grpc.NewClient(deadAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	node.peers = map[string]raftpb.RaftServiceClient{
		deadAddr: raftpb.NewRaftServiceClient(conn),
	}

	_, err = node.sendRequestVote(deadAddr, &raftpb.RequestVoteRequest{})
	if err == nil {
		t.Error("want error for unreachable peer, got nil")
	}
}

func TestRPC_Timeout(t *testing.T) {
	// slow server
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	slowAddr := ln.Addr().String()

	srv := grpc.NewServer()
	raftpb.RegisterRaftServiceServer(srv, &slowRaftService{})
	go srv.Serve(ln)
	t.Cleanup(srv.GracefulStop)
	conn, err := grpc.NewClient(slowAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { conn.Close() })

	// regular node
	node, _ := newUnstartedTestNode(t)

	node.peers = map[string]raftpb.RaftServiceClient{
		slowAddr: raftpb.NewRaftServiceClient(conn),
	}

	resp, err := node.sendRequestVote(slowAddr, &raftpb.RequestVoteRequest{})

	if err == nil {
		t.Errorf("want error, got resp: %v", resp)
	}
	if status.Code(err) != codes.DeadlineExceeded {
		t.Errorf("want DeadlineExceeded, got: %v", err)
	}
}
