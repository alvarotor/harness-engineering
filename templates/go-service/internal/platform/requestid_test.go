package platform

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_GeneratesWhenMissing(t *testing.T) {
	var seen string
	h := RequestID(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if seen == "" || rec.Header().Get(RequestIDHeader) != seen {
		t.Errorf("context id %q, response header %q; want the same generated ID", seen, rec.Header().Get(RequestIDHeader))
	}
}

func TestRequestID_PropagatesValidIncomingID(t *testing.T) {
	var seen string
	h := RequestID(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		seen = RequestIDFrom(r.Context())
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "upstream-42")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if seen != "upstream-42" || rec.Header().Get(RequestIDHeader) != "upstream-42" {
		t.Errorf("context id %q, response header %q; want upstream-42", seen, rec.Header().Get(RequestIDHeader))
	}
}

func TestRequestID_ReplacesUnsafeIncomingID(t *testing.T) {
	h := RequestID(slog.New(slog.DiscardHandler))(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "evil\nline")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if got := rec.Header().Get(RequestIDHeader); got == "" || strings.Contains(got, "evil") {
		t.Errorf("response id %q, want a freshly generated one", got)
	}
}

func TestRequestID_LoggerCarriesRequestID(t *testing.T) {
	var buf bytes.Buffer
	base := slog.New(slog.NewJSONHandler(&buf, nil))
	h := RequestID(base)(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		LoggerFrom(r.Context()).Info("handled")
	}))
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set(RequestIDHeader, "abc-123")
	h.ServeHTTP(httptest.NewRecorder(), req)
	if !strings.Contains(buf.String(), `"request_id":"abc-123"`) {
		t.Errorf("log line %q lacks request_id", buf.String())
	}
}

func TestRequestContext_OutsideRequest(t *testing.T) {
	ctx := context.Background()
	if RequestIDFrom(ctx) != "" {
		t.Error("RequestIDFrom outside a request must be empty")
	}
	if LoggerFrom(ctx) != slog.Default() {
		t.Error("LoggerFrom outside a request must return the default logger")
	}
}

func TestValidRequestID(t *testing.T) {
	cases := map[string]bool{
		"abc-DEF_1.2":            true,
		"":                       false,
		"a b":                    false,
		"a\nb":                   false,
		"é":                      false,
		strings.Repeat("a", 128): true,
		strings.Repeat("a", 129): false,
	}
	for id, want := range cases {
		if got := validRequestID(id); got != want {
			t.Errorf("validRequestID(%q) = %v, want %v", id, got, want)
		}
	}
}
