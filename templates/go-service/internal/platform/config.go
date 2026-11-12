// Package platform implements the runtime contract every deployable follows (ARCHITECTURE.md §8, §9).
package platform

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the validated environment of the service.
type Config struct {
	ServiceName     string
	Version         string
	Env             string
	LogLevel        slog.Level
	AppPort         int
	OpsPort         int
	ShutdownTimeout time.Duration
}

// maxShutdownTimeout is the runtime contract's limit: exit within 20 s of SIGTERM.
const maxShutdownTimeout = 20 * time.Second

var levels = map[string]slog.Level{"debug": slog.LevelDebug, "info": slog.LevelInfo, "warn": slog.LevelWarn, "error": slog.LevelError}

// LoadConfig reads and validates the environment; getenv is os.LookupEnv in production.
// It fails fast and names the variable. Every variable NAME can also be given as NAME_FILE (file secrets).
func LoadConfig(getenv func(string) (string, bool)) (Config, error) {
	r := &reader{getenv: getenv}
	c := Config{
		ServiceName:     r.get("SERVICE_NAME", "__NAME__"),
		Version:         r.get("VERSION", "dev"),
		Env:             r.get("ENV", "local"),
		LogLevel:        r.level("LOG_LEVEL", "info"),
		AppPort:         r.port("APP_PORT", 8080),
		OpsPort:         r.port("OPS_PORT", 8081),
		ShutdownTimeout: r.duration("SHUTDOWN_TIMEOUT", "20s"),
	}
	if r.err == nil && c.AppPort == c.OpsPort {
		r.err = errors.New("APP_PORT and OPS_PORT must differ")
	}
	return c, r.err
}

// reader keeps the first error so LoadConfig reads top to bottom without repeated checks.
type reader struct {
	getenv func(string) (string, bool)
	err    error
}

// get returns NAME, or the file named by NAME_FILE (trailing newlines removed), or def. Empty means unset.
func (r *reader) get(name, def string) string {
	if r.err != nil {
		return def
	}
	v, _ := r.getenv(name)
	path, _ := r.getenv(name + "_FILE")
	if path != "" {
		if v != "" {
			r.err = fmt.Errorf("set %s or %s_FILE, not both", name, name)
			return def
		}
		b, err := os.ReadFile(path) //nolint:gosec // the operator names this file through the environment
		if err != nil {
			r.err = fmt.Errorf("%s_FILE: %w", name, err)
			return def
		}
		v = strings.TrimRight(string(b), "\r\n")
	}
	if v == "" {
		return def
	}
	return v
}

func (r *reader) level(name, def string) slog.Level {
	s := strings.ToLower(r.get(name, def))
	l, ok := levels[s]
	if !ok && r.err == nil {
		r.err = fmt.Errorf("%s must be debug, info, warn or error, got %q", name, s)
	}
	return l
}

func (r *reader) port(name string, def int) int {
	s := r.get(name, strconv.Itoa(def))
	n, err := strconv.Atoi(s)
	if (err != nil || n < 1 || n > 65535) && r.err == nil {
		r.err = fmt.Errorf("%s must be a port from 1 to 65535, got %q", name, s)
	}
	return n
}

func (r *reader) duration(name, def string) time.Duration {
	s := r.get(name, def)
	d, err := time.ParseDuration(s)
	if (err != nil || d <= 0 || d > maxShutdownTimeout) && r.err == nil {
		r.err = fmt.Errorf("%s must be a Go duration above 0 and at most %s, got %q", name, maxShutdownTimeout, s)
	}
	return d
}
