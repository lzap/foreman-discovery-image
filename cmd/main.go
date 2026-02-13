package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"fdi/internal/facts"
	"fdi/internal/opt"
)

// TODO: add automatic detection of proxy/foreman
const SatelliteURL = "https://foreman.routed.lan/api/v2/discovered_hosts/facts"

func RegisterWithSatellite() error {
	var f facts.Facts
	if err := facts.Collect(&f, opt.Debug, opt.CustomPath); err != nil {
		return fmt.Errorf("failed to collect facts: %w", err)
	}

	payload := map[string]any{
		"facts": f,
	}

	jsonData, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode facts: %w", err)
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   30 * time.Second,
	}

	req, err := http.NewRequest("POST", SatelliteURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		fmt.Println("Successfully registered with Satellite!")
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	return fmt.Errorf("satellite returned error: %s: %s", resp.Status, string(body))
}

func main() {
	flag.Parse()

	if opt.Facts {
		var f facts.Facts
		if err := facts.Collect(&f, opt.Debug, opt.CustomPath); err != nil {
			log.Fatalf("failed to collect facts: %v", err)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]any{"facts": f}); err != nil {
			log.Fatalf("failed to encode facts: %v", err)
		}
		return
	}

	err := RegisterWithSatellite()
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
	}
}
