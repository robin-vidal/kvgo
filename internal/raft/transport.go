package raft

import (
	"context"

	"github.com/robin-vidal/kvgo/internal/raft/raftpb"
)

type grpcTransport struct {
	raftpb.UnimplementedRaftServiceServer
	node *Node
}

func (t *grpcTransport) RequestVote(ctx context.Context, req *raftpb.RequestVoteRequest) (*raftpb.RequestVoteResponse, error) {
	t.node.state.mu.RLock()
	currentTerm := t.node.state.CurrentTerm
	votedFor := t.node.state.VotedFor
	t.node.state.mu.RUnlock()

	if req.Term < currentTerm {
		return &raftpb.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	if req.Term > currentTerm {
		currentTerm = req.Term
		votedFor = ""
		t.node.state.SetTermAndVote(req.Term, votedFor)
	}

	if votedFor != "" && votedFor != req.CandidateId {
		return &raftpb.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	myTerm := t.node.wal.GetTerm()
	myIndex := t.node.wal.CurrentSeqNum()
	if req.LastLogTerm < myTerm || (req.LastLogTerm == myTerm && req.LastLogIndex < myIndex) {
		return &raftpb.RequestVoteResponse{
			Term:        currentTerm,
			VoteGranted: false,
		}, nil
	}

	t.node.state.SetTermAndVote(req.Term, req.CandidateId)

	t.node.mu.Lock()
	t.node.role = Follower
	t.node.mu.Unlock()

	return &raftpb.RequestVoteResponse{
		Term:        req.Term,
		VoteGranted: true,
	}, nil
}

func (t *grpcTransport) AppendEntries(ctx context.Context, req *raftpb.AppendEntriesRequest) (*raftpb.AppendEntriesResponse, error) {
	return &raftpb.AppendEntriesResponse{}, nil
}
