package config

import (
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name                    string
		args                    []string
		wantHost                string
		wantPort                int
		wantShardAmount         int
		wantDebug               bool
		wantWalPath             string
		wantSnapshotPath        string
		wantCompactionThreshold int
		wantNodeID              string
		wantPeers               []string
		wantElectionTimeoutMin  time.Duration
		wantElectionTimeoutMax  time.Duration
		wantHeartbeatInterval   time.Duration
		wantRaftStatePath       string
		wantErr                 bool
	}{
		{
			name:                    "default values",
			args:                    []string{"--nodeID=node-test"},
			wantHost:                "0.0.0.0",
			wantPort:                6379,
			wantShardAmount:         runtime.NumCPU(),
			wantDebug:               false,
			wantWalPath:             "/var/lib/kvgo/wal.log",
			wantSnapshotPath:        "/var/lib/kvgo/snapshot.db",
			wantCompactionThreshold: 10000,
			wantNodeID:              "node-test",
			wantPeers:               []string{},
			wantElectionTimeoutMin:  150 * time.Millisecond,
			wantElectionTimeoutMax:  300 * time.Millisecond,
			wantHeartbeatInterval:   50 * time.Millisecond,
			wantRaftStatePath:       "/var/lib/kvgo/raft.state",
			wantErr:                 false,
		},
		{
			name:                    "zero shard",
			args:                    []string{"--nodeID=node-test", "--shardAmount=0"},
			wantHost:                "0.0.0.0",
			wantPort:                6379,
			wantShardAmount:         0,
			wantDebug:               false,
			wantWalPath:             "/var/lib/kvgo/wal.log",
			wantSnapshotPath:        "/var/lib/kvgo/snapshot.db",
			wantCompactionThreshold: 10000,
			wantNodeID:              "node-test",
			wantPeers:               []string{},
			wantElectionTimeoutMin:  150 * time.Millisecond,
			wantElectionTimeoutMax:  300 * time.Millisecond,
			wantHeartbeatInterval:   50 * time.Millisecond,
			wantRaftStatePath:       "/var/lib/kvgo/raft.state",
			wantErr:                 true,
		},
		{
			name:                    "custom values",
			args:                    []string{"--nodeID=node-test", "--host=1.1.1.1", "--port=9999", "--debug=true", "--walPath=./wal.log", "--snapshotPath=./snapshot.db", "--compactionThreshold=1"},
			wantHost:                "1.1.1.1",
			wantPort:                9999,
			wantShardAmount:         runtime.NumCPU(),
			wantDebug:               true,
			wantWalPath:             "./wal.log",
			wantSnapshotPath:        "./snapshot.db",
			wantCompactionThreshold: 1,
			wantNodeID:              "node-test",
			wantPeers:               []string{},
			wantElectionTimeoutMin:  150 * time.Millisecond,
			wantElectionTimeoutMax:  300 * time.Millisecond,
			wantHeartbeatInterval:   50 * time.Millisecond,
			wantRaftStatePath:       "/var/lib/kvgo/raft.state",
			wantErr:                 false,
		},
		{
			name:                    "multiple peers",
			args:                    []string{"--nodeID=node-test", "--peers=node1:6379,node2:6379"},
			wantHost:                "0.0.0.0",
			wantPort:                6379,
			wantShardAmount:         runtime.NumCPU(),
			wantDebug:               false,
			wantWalPath:             "/var/lib/kvgo/wal.log",
			wantSnapshotPath:        "/var/lib/kvgo/snapshot.db",
			wantCompactionThreshold: 10000,
			wantNodeID:              "node-test",
			wantPeers:               []string{"node1:6379", "node2:6379"},
			wantElectionTimeoutMin:  150 * time.Millisecond,
			wantElectionTimeoutMax:  300 * time.Millisecond,
			wantHeartbeatInterval:   50 * time.Millisecond,
			wantRaftStatePath:       "/var/lib/kvgo/raft.state",
			wantErr:                 false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := Parse(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Parse() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				if cfg.Host != tt.wantHost {
					t.Errorf("Parse() error = %v, wantHost %v", cfg.Host, tt.wantHost)
				}
				if cfg.Port != tt.wantPort {
					t.Errorf("Parse() error = %v, wantPort %v", cfg.Port, tt.wantPort)
				}
				if cfg.Debug != tt.wantDebug {
					t.Errorf("Parse() error = %v, wantDebug %v", cfg.Debug, tt.wantDebug)
				}
				if cfg.ShardAmount != tt.wantShardAmount {
					t.Errorf("Parse() error = %v, wantShardAmount %v", cfg.ShardAmount, tt.wantShardAmount)
				}
				if cfg.WalConfig.WalPath != tt.wantWalPath {
					t.Errorf("Parse() error = %v, wantWallPath %v", cfg.WalConfig.WalPath, tt.wantWalPath)
				}
				if cfg.WalConfig.SnapshotPath != tt.wantSnapshotPath {
					t.Errorf("Parse() error = %v, wantSnapshotPath %v", cfg.WalConfig.SnapshotPath, tt.wantSnapshotPath)
				}
				if cfg.WalConfig.CompactionThreshold != tt.wantCompactionThreshold {
					t.Errorf("Parse() error = %v, wantCompactionThreshold %v", cfg.WalConfig.CompactionThreshold, tt.wantCompactionThreshold)
				}
				if cfg.RaftConfig.NodeID != tt.wantNodeID {
					t.Errorf("Parse() error = %v, wantNodeID %v", cfg.RaftConfig.NodeID, tt.wantNodeID)
				}
				if cfg.RaftConfig.ElectionTimeoutMin != tt.wantElectionTimeoutMin {
					t.Errorf("Parse() error = %v, wantElectionTimeoutMin %v", cfg.RaftConfig.ElectionTimeoutMin, tt.wantElectionTimeoutMin)
				}
				if cfg.RaftConfig.ElectionTimeoutMax != tt.wantElectionTimeoutMax {
					t.Errorf("Parse() error = %v, wantElectionTimeoutMax %v", cfg.RaftConfig.ElectionTimeoutMax, tt.wantElectionTimeoutMax)
				}
				if cfg.RaftConfig.HeartbeatInterval != tt.wantHeartbeatInterval {
					t.Errorf("Parse() error = %v, wantHeartbeatInterval %v", cfg.RaftConfig.HeartbeatInterval, tt.wantHeartbeatInterval)
				}
				if cfg.RaftConfig.RaftStatePath != tt.wantRaftStatePath {
					t.Errorf("Parse() error = %v, wantRaftStatePath %v", cfg.RaftConfig.RaftStatePath, tt.wantRaftStatePath)
				}
				if !slices.Equal(cfg.RaftConfig.Peers, tt.wantPeers) {
					t.Errorf("Parse() error = %v, wantPeers %v", strings.Join(cfg.RaftConfig.Peers, ","), strings.Join(tt.wantPeers, ","))
				}
			}
		})
	}
}
