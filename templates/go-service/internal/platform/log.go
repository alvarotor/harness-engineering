package platform

import (
	"io"
	"log/slog"
)

// NewLogger returns the JSON logger for stdout. Every line carries the runtime-contract attributes.
func NewLogger(w io.Writer, cfg Config) *slog.Logger {
	h := slog.NewJSONHandler(w, &slog.HandlerOptions{
		Level: cfg.LogLevel,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey && len(groups) == 0 {
				a.Key = "ts"
			}
			return a
		},
	})
	return slog.New(h).With("service", cfg.ServiceName, "version", cfg.Version, "env", cfg.Env)
}
