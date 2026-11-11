package platform

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"sync/atomic"
	"time"
)

// Timeouts bound slow clients (slowloris) and idle keep-alives. Defaults suit JSON APIs; raise them per deployable only with a reason.
const (
	readHeaderTimeout = 10 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
)

// newServer returns an http.Server with every timeout set.
func newServer(h http.Handler) *http.Server {
	return &http.Server{
		Handler:           h,
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
}

// Run listens on APP_PORT and OPS_PORT, then serves until ctx is cancelled (main wires SIGTERM/SIGINT) and drains.
func Run(ctx context.Context, cfg Config, log *slog.Logger, app http.Handler) error {
	appLn, err := listen(cfg.AppPort)
	if err != nil {
		return err
	}
	opsLn, err := listen(cfg.OpsPort)
	if err != nil {
		_ = appLn.Close()
		return err
	}
	return Serve(ctx, cfg, log, appLn, opsLn, app)
}

// Serve runs the app and ops servers on the given listeners. Cancelling ctx sets readiness to false, then
// drains both servers within cfg.ShutdownTimeout. The ops server stays up during the drain so /readyz reports 503.
func Serve(ctx context.Context, cfg Config, log *slog.Logger, app, ops net.Listener, appHandler http.Handler) error {
	var ready atomic.Bool
	ready.Store(true)
	appSrv := newServer(appHandler)
	opsSrv := newServer(opsHandler(&ready))

	failed := make(chan error, 2)
	go func() { failed <- appSrv.Serve(app) }()
	go func() { failed <- opsSrv.Serve(ops) }()

	var runErr error
	select {
	case runErr = <-failed: // a server stopped on its own; report it after the drain
	case <-ctx.Done():
		log.Info("shutdown started")
	}
	ready.Store(false)

	sctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	shutdownErr := errors.Join(appSrv.Shutdown(sctx), opsSrv.Shutdown(sctx))
	return errors.Join(runErr, shutdownErr)
}

func listen(port int) (net.Listener, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, fmt.Errorf("listen on port %d: %w", port, err)
	}
	return ln, nil
}

// opsHandler serves the probes. It is never routed publicly (runtime contract).
func opsHandler(ready *atomic.Bool) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /livez", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		status := http.StatusOK
		if !ready.Load() {
			status = http.StatusServiceUnavailable
		}
		w.WriteHeader(status)
	})
	return mux
}
