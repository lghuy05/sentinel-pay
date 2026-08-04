package logging

import (
	"log/slog"
	"testing"
)

func TestParseLevelDefaultsToInfo(t *testing.T) {
	if got := parseLevel(""); got != slog.LevelInfo {
		t.Fatalf("level = %v", got)
	}
}

func TestParseLevelAcceptsAliases(t *testing.T) {
	if got := parseLevel("warning"); got != slog.LevelWarn {
		t.Fatalf("level = %v", got)
	}
}
