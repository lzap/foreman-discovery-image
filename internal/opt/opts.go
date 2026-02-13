package opt

import "flag"

var (
	Facts       bool   // -facts: only print collected facts to stdout
	Debug       bool   // -debug: safe mode (no reboots/config) and print debugging information to stderr
	CustomPath  string // -custom_path: path to search for fdi-fact-* executables (default /usr/local/bin)
)

func init() {
	flag.BoolVar(&Facts, "facts", false, "only print collected facts to stdout")
	flag.BoolVar(&Debug, "debug", false, "safe mode and print debugging information to stderr")
	flag.StringVar(&CustomPath, "custom_path", "/usr/local/bin", "path to search for fdi-fact-* executables")
}
