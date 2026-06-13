package config

import "flag"

type WalConfig struct {
	WalPath string
}

func (cfg *WalConfig) Parse(fs *flag.FlagSet) {
	fs.StringVar(&cfg.WalPath, "walPath", "/var/lib/kvgo/wal.log", "Path to the WAL log file")
}
