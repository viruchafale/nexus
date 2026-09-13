// Package logger provides the shared stderr logger for NEXUS.
//
// Keep it thin: standard library log/slog only, no extra dependencies.
package logger

import (
	"log/slog"
	"os"
)

// New returns a text logger writing to stderr.
// verbose selects debug level, otherwise info level.
func New(verbose bool) *slog.Logger {
	level := slog.LevelInfo
	if verbose {
		level = slog.LevelDebug
	}
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: level,
	}))
}
