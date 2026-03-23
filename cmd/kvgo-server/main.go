/*
kvgo-server is the main entry point for the KVGo database engine.
It handles CLI flags, initializes the storage, and starts the TCP server.
*/
package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/database"
	"github.com/robin-vidal/kvgo/internal/logger"
	"github.com/robin-vidal/kvgo/internal/server"
	"github.com/robin-vidal/kvgo/internal/telemetry"
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

	db := database.New(cfg)

	err = server.Start(cfg, db)
	if err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
