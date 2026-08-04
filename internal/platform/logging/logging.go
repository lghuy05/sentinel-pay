package logging

import (
	"log/slog"
	"os"
	"strings"
)

func New(serviceName, level string) *slog.Logger {
	opts := &slog.HandlerOptions{Level: parseLevel(level)}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler).With("service", serviceName)
}

func parseLevel(level string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
