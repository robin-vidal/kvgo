// Package logger handles global logger initialisation
package logger

import (
	"log/slog"
	"os"

	"github.com/rvHoney/kvgo/internal/config"
)

func Init(cfg *config.Config) {
	level := slog.LevelInfo
	if cfg.Debug {
		level = slog.LevelDebug
	}

	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	})

	slog.SetDefault(slog.New(handler))
}
