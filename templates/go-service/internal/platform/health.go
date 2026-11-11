package platform

import (
	"fmt"
	"net/http"
	"time"
)

// Healthcheck probes the liveness URL. main exposes it as `service healthcheck` because the image has no shell.
func Healthcheck(url string) error {
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("%s returned %s", url, resp.Status)
	}
	return nil
}
