package config

import (
	"errors"
	"flag"
	"strings"
	"time"
)

type RaftConfig struct {
	NodeID   string
	Host     string
	Port     int
	Peers    []string
	peersRaw string

	ElectionTimeoutMin time.Duration
	ElectionTimeoutMax time.Duration
	HeartbeatInterval  time.Duration

	RaftStatePath string
}

func (cfg *RaftConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&cfg.NodeID, "nodeID", "", "Unique ID of this Raft node")
	fs.StringVar(&cfg.peersRaw, "peers", "", "Comma-separated list of peer addresses (e.g. node1:6379,node2:6379)")
	fs.StringVar(&cfg.Host, "raftHost", "0.0.0.0", "Host for Raft peer communication")
	fs.IntVar(&cfg.Port, "raftPort", 6380, "Port for Raft peer communication")

	fs.DurationVar(&cfg.ElectionTimeoutMin, "electionTimeoutMin", 150*time.Millisecond, "Minimum election timeout")
	fs.DurationVar(&cfg.ElectionTimeoutMax, "electionTimeoutMax", 300*time.Millisecond, "Maximum election timeout")
	fs.DurationVar(&cfg.HeartbeatInterval, "heartbeatInterval", 50*time.Millisecond, "Leader heartbeat interval")

	fs.StringVar(&cfg.RaftStatePath, "raftStatePath", "/var/lib/kvgo/raft.state", "Path to the Raft persistent state file")
}

func (cfg *RaftConfig) PostParse() error {
	if cfg.peersRaw != "" {
		cfg.Peers = strings.Split(cfg.peersRaw, ",")
		for i := range cfg.Peers {
			cfg.Peers[i] = strings.TrimSpace(cfg.Peers[i])
		}
	}
	return cfg.validate()
}

func (cfg *RaftConfig) validate() error {
	if cfg.NodeID == "" {
		return errors.New("config: nodeID is required")
	}
	if cfg.ElectionTimeoutMin >= cfg.ElectionTimeoutMax {
		return errors.New("config: electionTimeoutMin must be strictly less than electionTimeoutMax")
	}
	return nil
}
