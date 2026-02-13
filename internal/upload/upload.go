package upload

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"fdi/internal/facts"
)

// Endpoint is a facts upload endpoint (Foreman or Smart Proxy).
type Endpoint struct {
	URL  string // base URL (e.g. https://foreman.example.com)
	Type string // "foreman" or "proxy"
}

// factsURL returns the full URL to POST facts to.
func (e *Endpoint) factsURL() string {
	base := strings.TrimSuffix(e.URL, "/")
	if e.Type == "proxy" {
		return base + "/discovery/create"
	}
	return base + "/api/v2/discovered_hosts/facts"
}

// Once performs a single facts upload. Returns nil on success.
func (e *Endpoint) Once(f *facts.Facts) error {
	payload := map[string]any{"facts": f}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode facts: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	req, err := http.NewRequest("POST", e.factsURL(), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("endpoint returned %s: %s", resp.Status, string(b))
}
