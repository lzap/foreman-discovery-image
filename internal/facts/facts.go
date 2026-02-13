package facts

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Facts holds all facts
type Facts struct {
	DiscoveryBootIf MAC        `json:"discovery_bootif"`
	Interfaces      string     `json:"interfaces"`
	Networking      Networking `json:"networking"`
	SerialNumber    string     `json:"serialnumber,omitempty"`
	ProcessorCount  int        `json:"processorcount,omitempty"`
	Architecture    string     `json:"architecture,omitempty"`
	MemorySize      int        `json:"memorysize_mb,omitempty"` // MiB
	Memory          Memory     `json:"memory,omitempty"`
	Processors      Processors `json:"processors,omitempty"`
	DMI             DMI        `json:"dmi,omitempty"`
	Disks           Disks      `json:"disks,omitempty"`

	Custom map[string]any `json:"-"`
}

type Memory struct {
	System MemorySystem `json:"system,omitempty"`
}

type MemorySystem struct {
	TotalBytes int64 `json:"total_bytes,omitempty"`
}

type Processors struct {
	Count         int      `json:"count,omitempty"`
	PhysicalCount int      `json:"physicalcount,omitempty"`
	Models        []string `json:"models,omitempty"`
}

type DMI struct {
	Manufacturer string     `json:"manufacturer,omitempty"`
	Product      DMIProduct `json:"product,omitempty"`
}

type DMIProduct struct {
	Name string `json:"name,omitempty"`
	UUID string `json:"uuid,omitempty"`
}

type Disks map[string]DiskInfo

type DiskInfo struct {
	SizeBytes int64  `json:"size_bytes,omitempty"`
	Vendor    string `json:"vendor,omitempty"`
	Model     string `json:"model,omitempty"`
}

func (f Facts) MarshalJSON() ([]byte, error) {
	type temp Facts
	b, err := json.Marshal(temp(f))
	if err != nil {
		return nil, err
	}

	var out map[string]any
	if err := json.Unmarshal(b, &out); err != nil {
		return nil, err
	}

	maps.Copy(out, f.Custom)
	return json.Marshal(out)
}

type Networking struct {
	FQDN       string               `json:"fqdn"`
	Hostname   string               `json:"hostname"`
	IP         string               `json:"ip"`
	MAC        MAC                  `json:"mac"`
	Netmask    string               `json:"netmask"`
	Network    string               `json:"network"`
	Primary    string               `json:"primary"`
	Interfaces map[string]Interface `json:"interfaces,omitempty"`
}

type Interface struct {
	IP        string    `json:"ip"`
	IP6       string    `json:"ip6"`
	MTU       int       `json:"mtu"`
	MAC       MAC       `json:"mac"`
	Netmask   string    `json:"netmask"`
	Network   string    `json:"network"`
	Bindings  []Binding `json:"bindings,omitempty"`
	Bindings6 []Binding `json:"bindings6,omitempty"`
}

type Binding struct {
	Address string `json:"address"`
	Netmask string `json:"netmask"`
	Network string `json:"network"`
}

// CollectorFunc is a function that populates one or more fields of Facts.
// Collectors register themselves via Register and are invoked when gathering facts.
type CollectorFunc func(result *Facts) error

type collector struct {
	name string
	fn   CollectorFunc
}

var collectors []collector

// Register adds a collector to the registry with the given name (e.g. "cpu", "dmi").
// It is typically called from init(). The name is used for debug timing output.
func Register(name string, f CollectorFunc) {
	collectors = append(collectors, collector{name: name, fn: f})
}

type timed struct {
	name string
	d    time.Duration
}

// Collect runs all registered collectors and fills result. Then runs executable fdi-fact-* scripts in PWD and customPath.
// When customPath is empty, /usr/local/bin is used. When debug is true, logs one line with the top 5 slowest (collectors + scripts).
// Returns the first error from any collector.
func Collect(result *Facts, debug bool, customPath string) error {
	var times []timed
	for _, c := range collectors {
		start := time.Now()
		err := c.fn(result)
		if err != nil {
			return err
		}
		if debug {
			times = append(times, timed{c.name, time.Since(start)})
		}
	}

	runFdiFactScripts(result, debug, &times, customPath)

	if debug && len(times) > 0 {
		sort.Slice(times, func(i, j int) bool { return times[i].d > times[j].d })
		n := 5
		if n > len(times) {
			n = len(times)
		}
		parts := make([]string, n)
		for i := 0; i < n; i++ {
			parts[i] = fmt.Sprintf("%s %v", times[i].name, times[i].d)
		}
		log.Printf("stats: %s", strings.Join(parts, ", "))
	}
	return nil
}

const fdiFactPrefix = "fdi-fact-"

func runFdiFactScripts(result *Facts, debug bool, times *[]timed, customPath string) {
	type script struct {
		path string
		key  string
	}
	if customPath == "" {
		customPath = "/usr/local/bin"
	}
	pwd, _ := os.Getwd()
	dirs := []string{pwd, customPath}
	if pwd == "" {
		dirs = []string{customPath}
	}
	seen := make(map[string]bool)
	var scripts []script
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			name := e.Name()
			if !strings.HasPrefix(name, fdiFactPrefix) {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			if info.Mode()&0111 == 0 {
				continue
			}
			key := name[len(fdiFactPrefix):]
			if key == "" || seen[key] {
				continue
			}
			seen[key] = true
			path := filepath.Join(dir, name)
			scripts = append(scripts, script{path, key})
		}
	}

	for _, s := range scripts {
		start := time.Now()
		cmd := exec.Command(s.path)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err := cmd.Run()
		d := time.Since(start)
		if debug {
			*times = append(*times, timed{"fdi-fact-" + s.key, d})
		}

		if result.Custom == nil {
			result.Custom = make(map[string]any)
		}
		result.Custom[s.key] = strings.TrimSpace(stdout.String())
		if t := strings.TrimSpace(stderr.String()); t != "" {
			result.Custom[s.key+"_stderr"] = t
		}
		exitCode := 0
		if err != nil {
			if exitErr, ok := err.(*exec.ExitError); ok {
				exitCode = exitErr.ExitCode()
			} else {
				exitCode = -1
			}
		}
		if exitCode != 0 {
			result.Custom[s.key+"_code"] = exitCode
		}
	}
}
