// Package server handles TCP communications
package server

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"time"

	"github.com/robin-vidal/kvgo/internal/command"
	"github.com/robin-vidal/kvgo/internal/config"
	"github.com/robin-vidal/kvgo/internal/resp"
	"github.com/robin-vidal/kvgo/internal/store"
)

// executeCommand dispatches the command based on its name and run it.
func executeCommand(s *store.Store, m *metrics, cmd resp.Command) []byte {
	res := command.Dispatch(s, cmd.Name, cmd.Args)
	m.recordCommand(res.CmdName, res.Status)
	return res.Response
}

// handleConnection manages a TCP connection, reading and executing commands in a loop.
func handleConnection(conn net.Conn, s *store.Store, m *metrics) {
	defer func() {
		if err := conn.Close(); err != nil {
			slog.Debug("failed to close connection", "error", err)
		}
		m.recordConnection(-1)
	}()
	slog.Debug("new TCP connection", "remoteAddr", conn.RemoteAddr())
	m.recordConnection(1)

	defer func() {
		if r := recover(); r != nil {
			slog.Error("panic in connection handler", "recover", r, "remoteAddr", conn.RemoteAddr())
		}
	}()

	reader := bufio.NewReader(conn)

	for {
		cmd, err := resp.ParseCommand(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("client disconnected", "remoteAddr", conn.RemoteAddr())
			} else {
				slog.Debug("packet parsing fail", "error", err)
			}
			break
		}

		start := time.Now()
		response := executeCommand(s, m, cmd)
		m.recordDuration(cmd.Name, float64(time.Since(start).Microseconds()))

		slog.Debug("executed", "cmd", cmd, "response", response)

		_, err = conn.Write(response)
		if err != nil {
			slog.Debug("failed to send response", "error", err)
			break
		}
	}
}

// Start launches a TCP server according to the configuration
func Start(cfg *config.Config, s *store.Store) error {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	defer func() {
		if err := ln.Close(); err != nil {
			slog.Debug("failed to close listener", "error", err)
		}
	}()

	slog.Info("TCP server is listening", "addr", ln.Addr().String())

	m, err := newMetrics(s)
	if err != nil {
		return err
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			slog.Debug("connection accept failed", "error", err)
			continue
		}

		go handleConnection(conn, s, m)
	}
}
