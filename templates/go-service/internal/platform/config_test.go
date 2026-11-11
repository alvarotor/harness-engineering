package platform

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func mapEnv(m map[string]string) func(string) (string, bool) {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoadConfig_DefaultsWhenUnset(t *testing.T) {
	got, err := LoadConfig(mapEnv(nil))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{ServiceName: "__NAME__", Version: "dev", Env: "local", LogLevel: slog.LevelInfo, AppPort: 8080, OpsPort: 8081, ShutdownTimeout: 20 * time.Second}
	if got != want {
		t.Errorf("config = %+v, want %+v", got, want)
	}
}

func TestLoadConfig_ReadsEveryVariable(t *testing.T) {
	got, err := LoadConfig(mapEnv(map[string]string{
		"SERVICE_NAME": "orders", "VERSION": "1.2.3", "ENV": "dev", "LOG_LEVEL": "WARN",
		"APP_PORT": "9000", "OPS_PORT": "9001", "SHUTDOWN_TIMEOUT": "5s",
	}))
	if err != nil {
		t.Fatal(err)
	}
	want := Config{ServiceName: "orders", Version: "1.2.3", Env: "dev", LogLevel: slog.LevelWarn, AppPort: 9000, OpsPort: 9001, ShutdownTimeout: 5 * time.Second}
	if got != want {
		t.Errorf("config = %+v, want %+v", got, want)
	}
}

func TestLoadConfig_RejectsInvalidValues(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"unknown log level", map[string]string{"LOG_LEVEL": "trace"}, "LOG_LEVEL"},
		{"non-numeric app port", map[string]string{"APP_PORT": "http"}, "APP_PORT"},
		{"zero ops port", map[string]string{"OPS_PORT": "0"}, "OPS_PORT"},
		{"port above range", map[string]string{"APP_PORT": "70000"}, "APP_PORT"},
		{"same app and ops port", map[string]string{"APP_PORT": "8081"}, "must differ"},
		{"unparsable timeout", map[string]string{"SHUTDOWN_TIMEOUT": "soon"}, "SHUTDOWN_TIMEOUT"},
		{"zero timeout", map[string]string{"SHUTDOWN_TIMEOUT": "0s"}, "SHUTDOWN_TIMEOUT"},
		{"timeout above contract", map[string]string{"SHUTDOWN_TIMEOUT": "21s"}, "SHUTDOWN_TIMEOUT"},
		{"file with bad value", map[string]string{"OPS_PORT_FILE": writeFile(t, "nope")}, "OPS_PORT"},
		{"value and file both set", map[string]string{"VERSION": "1", "VERSION_FILE": writeFile(t, "2")}, "VERSION"},
		{"missing file", map[string]string{"ENV_FILE": filepath.Join(t.TempDir(), "absent")}, "ENV_FILE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadConfig(mapEnv(tc.env))
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Errorf("err = %v, want one naming %q", err, tc.want)
			}
		})
	}
}

func TestLoadConfig_ReadsFileVariants(t *testing.T) {
	got, err := LoadConfig(mapEnv(map[string]string{
		"APP_PORT_FILE":     writeFile(t, "9100\r\n"),
		"SERVICE_NAME_FILE": writeFile(t, "billing\n"),
		"ENV_FILE":          writeFile(t, ""), // empty file means unset
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.AppPort != 9100 || got.ServiceName != "billing" || got.Env != "local" {
		t.Errorf("config = %+v", got)
	}
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "value")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}
