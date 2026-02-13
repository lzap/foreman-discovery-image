package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"fdi/internal/facts"
	"fdi/internal/opt"
	"fdi/internal/upload"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	typeStr := opt.Type

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

	if opt.Once {
		var f facts.Facts
		if err := facts.Collect(&f, opt.Debug, opt.CustomPath); err != nil {
			log.Fatalf("failed to collect facts: %v", err)
		}
		if opt.URL == "" {
			log.Fatalf("-url or proxy.url is required for -once")
		}
		ep := &upload.Endpoint{URL: opt.URL, Type: typeStr}
		if err := ep.Once(&f); err != nil {
			log.Fatalf("upload failed: %v", err)
		}
		if opt.Debug {
			log.Println("Successful fact upload")
		}
		return
	}

	if opt.URL == "" {
		log.Fatalf("endpoint URL required for service mode (set -url or proxy.url on kernel command line)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		ep := &upload.Endpoint{URL: opt.URL, Type: typeStr}
		ep.Loop(ctx, time.Duration(opt.UploadSleep)*time.Second, opt.CustomPath, opt.Debug)
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	cancel()
	wg.Wait()
}
