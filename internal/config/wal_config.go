package config

import "flag"

type WalConfig struct {
	WalPath             string
	SnapshotPath        string
	CompactionThreshold int
}

func (cfg *WalConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&cfg.WalPath, "walPath", "/var/lib/kvgo/wal.log", "Path to the WAL log file")
	fs.StringVar(&cfg.SnapshotPath, "snapshotPath", "/var/lib/kvgo/snapshot.db", "Path to the snapshot file")
	fs.IntVar(&cfg.CompactionThreshold, "compactionThreshold", 10000, "Number of WAL entries before triggering compaction")
}
