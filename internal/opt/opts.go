package opt

import (
	"flag"
	"log"
	"strconv"

	"fdi/internal/cmdline"
)

var (
	Facts       bool   // -facts: only print collected facts to stdout
	Debug       bool   // -debug: safe mode (no reboots/config) and print debugging information to stderr
	CustomPath  string // -custom_path: path to search for fdi-fact-* executables (default /usr/local/bin)
	Once        bool   // -once: perform a single facts upload to the endpoint
	URL         string // -url: endpoint base URL (default from proxy.url)
	Type        string // -type: endpoint type (default from proxy.type or "foreman")
	UploadSleep int    // -uploadsleep: seconds between uploads in service mode (default from fdi.uploadsleep or DefaultUploadSleepSec)
)

func init() {
	defaultURL := cmdline.Get("proxy.url")
	defaultType := cmdline.GetDefault("proxy.type", "foreman")
	defaultUploadSleep, err := strconv.Atoi(cmdline.GetDefault("fdi.uploadsleep", "30"))

	if err != nil {
		log.Printf("unable to parse uploadsleep: %v", err)
		defaultUploadSleep = 30
	}

	flag.BoolVar(&Facts, "facts", false, "only print collected facts to stdout")
	flag.BoolVar(&Debug, "debug", false, "safe mode and print debugging information to stderr")
	flag.StringVar(&CustomPath, "custom_path", "/usr/local/bin", "path to search for fdi-fact-* executables")
	flag.BoolVar(&Once, "once", false, "perform a single facts upload to the endpoint")
	flag.StringVar(&URL, "url", defaultURL, "endpoint base URL (e.g. https://foreman.example.com)")
	flag.StringVar(&Type, "type", defaultType, "endpoint type: foreman or proxy")
	flag.IntVar(&UploadSleep, "uploadsleep", defaultUploadSleep, "seconds between fact uploads in service mode")
}
