package cmdline

import (
	"os"
	"strings"
)

var options map[string]string

func init() {
	options = parseKernelCmdline()
}

func parseKernelCmdline() map[string]string {
	out := make(map[string]string)
	data, err := os.ReadFile("/proc/cmdline")
	if err != nil {
		return out
	}
	for token := range strings.FieldsSeq(string(data)) {
		key, value, ok := strings.Cut(token, "=")
		if ok {
			out[key] = value
		} else {
			out[token] = ""
		}
	}
	return out
}

// Get returns the value for the given kernel command line option key.
// Returns empty string if the option is not set. Quoting is not supported.
func Get(key string) string {
	return options[key]
}
