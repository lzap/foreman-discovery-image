package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"fdi/internal/cmdline"
	"fdi/internal/facts"
	"fdi/internal/opt"
	"fdi/internal/upload"
)

func main() {
	flag.Parse()
	log.SetFlags(0)

	url := opt.URL
	if url == "" {
		url = cmdline.Get("proxy.url")
	}
	typeStr := opt.Type
	if typeStr == "" {
		typeStr = cmdline.Get("proxy.type")
	}
	if typeStr == "" {
		typeStr = "foreman"
	}

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
		if url == "" {
			log.Fatalf("-url or proxy.url is required for -once")
		}
		ep := &upload.Endpoint{URL: url, Type: typeStr}
		if err := ep.Once(&f); err != nil {
			log.Fatalf("upload failed: %v", err)
		}
		if opt.Debug {
			log.Println("Successful fact upload")
		}
		return
	}

	fmt.Fprintln(os.Stderr, "Service not implemented yet")
	os.Exit(1)
}
