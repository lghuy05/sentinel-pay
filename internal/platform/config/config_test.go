package config

import (
	"testing"
	"time"
)

func TestEnvStringUsesFallbackForMissingOrBlank(t *testing.T) {
	env := NewMapEnv(map[string]string{"BLANK": " "})

	if got := env.String("MISSING", "fallback"); got != "fallback" {
		t.Fatalf("missing value = %q", got)
	}
	if got := env.String("BLANK", "fallback"); got != "fallback" {
		t.Fatalf("blank value = %q", got)
	}
}

func TestEnvRequiredStringRejectsBlank(t *testing.T) {
	env := NewMapEnv(map[string]string{"DB_HOST": ""})

	if _, err := env.RequiredString("DB_HOST"); err == nil {
		t.Fatal("expected error for blank required value")
	}
}

func TestEnvParsesIntAndDuration(t *testing.T) {
	env := NewMapEnv(map[string]string{
		"PORT":    "8087",
		"TIMEOUT": "1500ms",
	})

	port, err := env.Int("PORT", 0)
	if err != nil {
		t.Fatal(err)
	}
	if port != 8087 {
		t.Fatalf("port = %d", port)
	}

	timeout, err := env.Duration("TIMEOUT", time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if timeout != 1500*time.Millisecond {
		t.Fatalf("timeout = %s", timeout)
	}
}
