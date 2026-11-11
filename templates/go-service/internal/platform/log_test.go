package platform

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"
)

func TestNewLogger_JSONWithContractAttributes(t *testing.T) {
	var buf bytes.Buffer
	cfg := Config{ServiceName: "svc", Version: "1.2.3", Env: "dev", LogLevel: slog.LevelInfo}
	log := NewLogger(&buf, cfg)
	log.Debug("hidden")
	log.Info("hello", "k", "v")

	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	if len(lines) != 1 {
		t.Fatalf("got %d lines, want 1 (debug must be filtered at info): %s", len(lines), buf.String())
	}
	var line map[string]any
	if err := json.Unmarshal(lines[0], &line); err != nil {
		t.Fatalf("not JSON: %v", err)
	}
	for k, want := range map[string]string{"service": "svc", "version": "1.2.3", "env": "dev", "msg": "hello", "k": "v"} {
		if line[k] != want {
			t.Errorf("%s = %v, want %q", k, line[k], want)
		}
	}
	if _, ok := line["ts"]; !ok {
		t.Errorf("missing ts in %s", lines[0])
	}
}
