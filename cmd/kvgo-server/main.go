/*
kvgo-server is the main entry point for the KVGo database engine.
It handles CLI flags, initializes the storage, and starts the TCP server.
*/
package main

import (
	"log/slog"
	"os"

	"github.com/rvHoney/kvgo/internal/config"
	"github.com/rvHoney/kvgo/internal/database"
	"github.com/rvHoney/kvgo/internal/logger"
	"github.com/rvHoney/kvgo/internal/server"
)

func main() {
	cfg, err := config.Parse(os.Args[1:])
	if err != nil {
		os.Exit(1)
	}

	logger.Init(cfg)

	db := database.New(cfg)

	err = server.Start(cfg, db)
	if err != nil {
		slog.Error("server stopped unexpectedly", "error", err)
		os.Exit(1)
	}
}
