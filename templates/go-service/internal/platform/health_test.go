package platform

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	ok := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer ok.Close()
	bad := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusServiceUnavailable) }))
	defer bad.Close()
	gone := httptest.NewServer(http.NotFoundHandler())
	goneURL := gone.URL
	gone.Close()

	if err := Healthcheck(ok.URL + "/livez"); err != nil {
		t.Errorf("200 must pass: %v", err)
	}
	if err := Healthcheck(bad.URL + "/livez"); err == nil {
		t.Error("503 must fail")
	}
	if err := Healthcheck(goneURL + "/livez"); err == nil {
		t.Error("unreachable server must fail")
	}
}
