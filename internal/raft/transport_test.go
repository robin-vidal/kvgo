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
	"google.golang.org/grpc/credentials/insecure"
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
		RPCTimeout:         100 * time.Millisecond,
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
