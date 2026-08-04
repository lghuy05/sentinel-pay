package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Env struct {
	lookup func(string) (string, bool)
}

func NewEnv() Env {
	return Env{lookup: os.LookupEnv}
}

func NewMapEnv(values map[string]string) Env {
	return Env{lookup: func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}}
}

func (e Env) String(key, fallback string) string {
	value, ok := e.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func (e Env) RequiredString(key string) (string, error) {
	value, ok := e.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("required environment variable %s is not set", key)
	}
	return value, nil
}

func (e Env) Int(key string, fallback int) (int, error) {
	value, ok := e.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}
	return parsed, nil
}

func (e Env) Duration(key string, fallback time.Duration) (time.Duration, error) {
	value, ok := e.lookup(key)
	if !ok || strings.TrimSpace(value) == "" {
		return fallback, nil
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as duration: %w", key, err)
	}
	return parsed, nil
}
