package opt

import "flag"

var (
	Facts      bool   // -facts: only print collected facts to stdout
	Debug      bool   // -debug: safe mode (no reboots/config) and print debugging information to stderr
	CustomPath string // -custom_path: path to search for fdi-fact-* executables (default /usr/local/bin)
	Once       bool   // -once: perform a single facts upload to the endpoint
	URL        string // -url: endpoint base URL (e.g. https://foreman.example.com)
	Type       string // -type: endpoint type ("foreman" or "proxy")
)

func init() {
	flag.BoolVar(&Facts, "facts", false, "only print collected facts to stdout")
	flag.BoolVar(&Debug, "debug", false, "safe mode and print debugging information to stderr")
	flag.StringVar(&CustomPath, "custom_path", "/usr/local/bin", "path to search for fdi-fact-* executables")
	flag.BoolVar(&Once, "once", false, "perform a single facts upload to the endpoint")
	flag.StringVar(&URL, "url", "", "endpoint base URL (e.g. https://foreman.example.com)")
	flag.StringVar(&Type, "type", "foreman", "endpoint type: foreman or proxy")
}
