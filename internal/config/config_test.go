package config

import (
	"runtime"
	"testing"
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
		wantErr                 bool
	}{
		{
			name:                    "default values",
			args:                    []string{},
			wantHost:                "0.0.0.0",
			wantPort:                6379,
			wantShardAmount:         runtime.NumCPU(),
			wantDebug:               false,
			wantWalPath:             "/var/lib/kvgo/wal.log",
			wantSnapshotPath:        "/var/lib/kvgo/snapshot.db",
			wantCompactionThreshold: 10000,
			wantErr:                 false,
		},
		{
			name:                    "zero shard",
			args:                    []string{"--shardAmount=0"},
			wantHost:                "0.0.0.0",
			wantPort:                6379,
			wantShardAmount:         0,
			wantDebug:               false,
			wantWalPath:             "/var/lib/kvgo/wal.log",
			wantSnapshotPath:        "/var/lib/kvgo/snapshot.db",
			wantCompactionThreshold: 10000,
			wantErr:                 true,
		},
		{
			name:                    "custom values",
			args:                    []string{"--host=1.1.1.1", "--port=9999", "--debug=true", "--walPath=./wal.log", "--snapshotPath=./snapshot.db", "--compactionThreshold=1"},
			wantHost:                "1.1.1.1",
			wantPort:                9999,
			wantShardAmount:         runtime.NumCPU(),
			wantDebug:               true,
			wantWalPath:             "./wal.log",
			wantSnapshotPath:        "./snapshot.db",
			wantCompactionThreshold: 1,
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
			}
		})
	}
}
