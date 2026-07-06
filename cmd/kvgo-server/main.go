/*
kvgo-server is the main entry point for the KVGo database engine.
It handles CLI flags, initializes the storage, and starts the TCP server.
*/
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/robin-vidal/kvgo/internal/command"
	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/logger"
	"github.com/robin-vidal/kvgo/internal/server"
	"github.com/robin-vidal/kvgo/internal/store"
	"github.com/robin-vidal/kvgo/internal/telemetry"
	"github.com/robin-vidal/kvgo/internal/wal"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	logger.Init(cfg)

	shutdown, err := telemetry.Init()
	if err != nil {
		slog.Error("failed to initialize telemetry", "error", err)
		os.Exit(1)
	}
	defer shutdown(context.Background())

	db, err := database.New(cfg)
	if err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}

	w, err := wal.Open(cfg.WalConfig)
	if err != nil {
		slog.Error("failed to open WAL", "error", err)
		os.Exit(1)
	}
	defer w.Close()

	snapshot, err := w.LoadSnapshot()
	if err != nil {
		slog.Error("failed to load snapshot", "error", err)
		os.Exit(1)
	}
	for k, v := range snapshot {
		db.Set(k, v)
	}

	entries, err := w.Replay()
	if err != nil {
		slog.Error("failed to replay WAL", "error", err)
		os.Exit(1)
	}

	for _, e := range entries {
		if err := command.Apply(db, e); err != nil {
			slog.Error("failed to apply WAL entry during replay", "error", err)
			os.Exit(1)
		}
	}

	s := store.New(db, w)

	err = server.Start(cfg, s)
	if err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
