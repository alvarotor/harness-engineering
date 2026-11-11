// Command service is the wiring entrypoint: config, logger, routes and lifecycle. Logic lives in internal/.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"har/services/__NAME__/internal/platform"
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	cfg, err := platform.LoadConfig(os.LookupEnv)
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		return 2
	}
	if len(args) == 1 && args[0] == "healthcheck" {
		if err := platform.Healthcheck(fmt.Sprintf("http://127.0.0.1:%d/livez", cfg.OpsPort)); err != nil {
			fmt.Fprintln(os.Stderr, "healthcheck:", err)
			return 1
		}
		return 0
	}
	if len(args) > 0 {
		fmt.Fprintln(os.Stderr, "usage: service [healthcheck]")
		return 2
	}

	log := platform.NewLogger(os.Stdout, cfg)
	slog.SetDefault(log)
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// No routes yet: every request is a 404 until the first feature adds them.
	app := platform.RequestID(log)(http.NewServeMux())
	if err := platform.Run(ctx, cfg, log, app); err != nil {
		log.Error("service stopped with an error", "err", err)
		return 1
	}
	return 0
}
