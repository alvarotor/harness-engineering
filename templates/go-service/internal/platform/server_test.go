package platform

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"testing"
	"time"
)

// Readiness is polled with a deadline; channels synchronize the rest. No fixed sleeps.

func testConfig() Config {
	return Config{ServiceName: "svc", ShutdownTimeout: 5 * time.Second}
}

func listenAny(t *testing.T) net.Listener {
	t.Helper()
	ln, err := net.Listen("tcp", ":0") //nolint:gosec // test listener on a free port; production binds all interfaces by design
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ln.Close() })
	return ln
}

func urlOf(ln net.Listener) string {
	return fmt.Sprintf("http://127.0.0.1:%d", ln.Addr().(*net.TCPAddr).Port)
}

func statusOf(url string) int {
	resp, err := http.Get(url) //nolint:gosec // test polls a local URL it built itself
	if err != nil {
		return 0
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

func waitFor(t *testing.T, what string, ok func() bool) {
	t.Helper()
	tick := time.NewTicker(5 * time.Millisecond)
	defer tick.Stop()
	timeout := time.After(5 * time.Second)
	for !ok() {
		select {
		case <-timeout:
			t.Fatalf("timed out waiting for %s", what)
		case <-tick.C:
		}
	}
}

func waitClosed(t *testing.T, ch <-chan struct{}, what string) {
	t.Helper()
	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for %s", what)
	}
}

func TestServe_ReadinessFlipsWhileDraining(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	app := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	})
	appLn, opsLn := listenAny(t), listenAny(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, testConfig(), slog.New(slog.DiscardHandler), appLn, opsLn, app) }()

	ops := urlOf(opsLn)
	waitFor(t, "readiness 200", func() bool { return statusOf(ops+"/readyz") == http.StatusOK })
	if statusOf(ops+"/livez") != http.StatusOK {
		t.Fatal("liveness must be 200 while running")
	}

	inflight := make(chan struct{})
	go func() {
		defer close(inflight)
		_ = statusOf(urlOf(appLn) + "/slow")
	}()
	waitClosed(t, entered, "in-flight request")

	cancel()
	waitFor(t, "readiness 503 during drain", func() bool { return statusOf(ops+"/readyz") == http.StatusServiceUnavailable })
	if statusOf(ops+"/livez") != http.StatusOK {
		t.Error("liveness must stay 200 while draining")
	}

	close(release)
	waitClosed(t, inflight, "in-flight request to finish")
	if err := <-done; err != nil {
		t.Fatalf("Serve returned %v after a clean drain", err)
	}
}

func TestServe_ReportsDrainTimeout(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	t.Cleanup(func() { close(release) })
	app := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		close(entered)
		<-release
	})
	appLn, opsLn := listenAny(t), listenAny(t)
	cfg := testConfig()
	cfg.ShutdownTimeout = 50 * time.Millisecond
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- Serve(ctx, cfg, slog.New(slog.DiscardHandler), appLn, opsLn, app) }()

	go func() { _ = statusOf(urlOf(appLn) + "/stuck") }()
	waitClosed(t, entered, "in-flight request")
	cancel()
	if err := <-done; err == nil {
		t.Fatal("Serve must report the drain deadline when a request outlives SHUTDOWN_TIMEOUT")
	}
}

func TestServe_ReturnsServerFailure(t *testing.T) {
	appLn, opsLn := listenAny(t), listenAny(t)
	_ = appLn.Close() // the app server cannot accept, so it fails on its own
	err := Serve(context.Background(), testConfig(), slog.New(slog.DiscardHandler), appLn, opsLn, http.NotFoundHandler())
	if err == nil {
		t.Fatal("a failed server must make Serve return an error")
	}
}

func TestRun_StartsAndStopsCleanly(t *testing.T) {
	cfg := testConfig() // ports 0: the OS picks free ports
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Run(ctx, cfg, slog.New(slog.DiscardHandler), http.NotFoundHandler()) }()
	cancel()
	if err := <-done; err != nil {
		t.Fatalf("Run returned %v after a clean stop", err)
	}
}

func TestRun_ReturnsListenErrors(t *testing.T) {
	busy := listenAny(t)
	port := busy.Addr().(*net.TCPAddr).Port
	cases := map[string]Config{
		"app port busy": {AppPort: port},
		"ops port busy": {OpsPort: port},
	}
	for name, cfg := range cases {
		t.Run(name, func(t *testing.T) {
			err := Run(context.Background(), cfg, slog.New(slog.DiscardHandler), http.NotFoundHandler())
			if err == nil {
				t.Fatal("Run must fail when a port is taken")
			}
		})
	}
}

func TestNewServer_SetsAllTimeouts(t *testing.T) {
	srv := newServer(http.NotFoundHandler())
	if srv.ReadHeaderTimeout != readHeaderTimeout || srv.ReadTimeout != readTimeout ||
		srv.WriteTimeout != writeTimeout || srv.IdleTimeout != idleTimeout {
		t.Errorf("timeouts = header %v read %v write %v idle %v; want the package defaults",
			srv.ReadHeaderTimeout, srv.ReadTimeout, srv.WriteTimeout, srv.IdleTimeout)
	}
	if srv.ReadHeaderTimeout <= 0 || srv.ReadTimeout <= 0 || srv.WriteTimeout <= 0 || srv.IdleTimeout <= 0 {
		t.Error("every timeout must be positive; zero means unlimited")
	}
}
