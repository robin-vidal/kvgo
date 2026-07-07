package config

import (
	"flag"
	"time"
)

type RaftConfig struct {
	NodeId   string
	Peers    []string
	peersRaw string

	ElectionTimeoutMin time.Duration
	ElectionTimeoutMax time.Duration
	HeartbeatInterval  time.Duration

	RaftStatePath string
}

func (cfg *RaftConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&cfg.NodeId, "nodeId", "", "Unique ID of this Raft node")
	fs.StringVar(&cfg.peersRaw, "peers", "", "Comma-separated list of peer addresses (e.g. node1:6379,node2:6379)")

	fs.DurationVar(&cfg.ElectionTimeoutMin, "electionTimeoutMin", 150*time.Millisecond, "Minimum election timeout")
	fs.DurationVar(&cfg.ElectionTimeoutMax, "electionTimeoutMax", 300*time.Millisecond, "Maximum election timeout")
	fs.DurationVar(&cfg.HeartbeatInterval, "heartbeatInterval", 50*time.Millisecond, "Leader heartbeat interval")

	fs.StringVar(&cfg.RaftStatePath, "raftStatePath", "/var/lib/kvgo/raft.state", "Path to the Raft persistent state file")
}
