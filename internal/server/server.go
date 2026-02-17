// Package server handles tcp communications
package server

import (
	"fmt"
	"log/slog"
	"net"

	"github.com/rvHoney/kvgo/internal/config"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()
	slog.Debug("new TCP connection", "remoteAddr", conn.RemoteAddr())
}

// Start launches a tcp server according to the configuration
func Start(cfg *config.Config) error {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer ln.Close()

	slog.Info("TCP server is listening", "addr", ln.Addr().String())
	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Debug("connection accept failed", "error", err)
			continue
		}
		go handleConnection(conn)
	}
}
