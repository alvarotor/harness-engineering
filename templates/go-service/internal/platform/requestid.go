package platform

import (
	"context"
	"crypto/rand"
	"log/slog"
	"net/http"
)

// RequestIDHeader carries the correlation ID across hops (runtime contract).
const RequestIDHeader = "X-Request-ID"

const maxRequestIDLen = 128

type requestKey struct{}

type requestScope struct {
	id  string
	log *slog.Logger
}

// RequestID reads or generates the request ID, echoes it in the response, and stores it with a
// per-request logger in the request context.
func RequestID(base *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := r.Header.Get(RequestIDHeader)
			if !validRequestID(id) {
				id = rand.Text()
			}
			w.Header().Set(RequestIDHeader, id)
			scope := requestScope{id: id, log: base.With("request_id", id)}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestKey{}, scope)))
		})
	}
}

// RequestIDFrom returns the request ID stored by RequestID, or "" outside a request.
func RequestIDFrom(ctx context.Context) string {
	s, _ := ctx.Value(requestKey{}).(requestScope)
	return s.id
}

// LoggerFrom returns the per-request logger, or the default logger outside a request.
func LoggerFrom(ctx context.Context) *slog.Logger {
	if s, ok := ctx.Value(requestKey{}).(requestScope); ok {
		return s.log
	}
	return slog.Default()
}

// validRequestID accepts 1-128 letters, digits, '-', '_' or '.'. Anything else is replaced, so untrusted input never reaches headers or logs.
func validRequestID(id string) bool {
	if id == "" || len(id) > maxRequestIDLen {
		return false
	}
	for i := range len(id) {
		c := id[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '_', c == '.':
		default:
			return false
		}
	}
	return true
}
