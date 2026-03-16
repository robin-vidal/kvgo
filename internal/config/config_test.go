package config

import (
	"runtime"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		wantHost        string
		wantPort        int
		wantShardAmount int
		wantDebug       bool
		wantErr         bool
	}{
		{
			name:            "Default values",
			args:            []string{},
			wantHost:        "localhost",
			wantPort:        6379,
			wantShardAmount: runtime.NumCPU(),
			wantDebug:       false,
			wantErr:         false,
		},
		{
			name:            "Zero Shard",
			args:            []string{"--shardAmount=0"},
			wantHost:        "localhost",
			wantPort:        6379,
			wantShardAmount: 0,
			wantDebug:       false,
			wantErr:         true,
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
			}
		})
	}
}
