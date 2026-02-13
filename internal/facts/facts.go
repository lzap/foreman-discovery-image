package facts

import (
	"encoding/json"
	"maps"
)

// Facts holds all facts
type Facts struct {
	DiscoveryBootIf string     `json:"discovery_bootif"`
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
	MAC        string               `json:"mac"`
	Netmask    string               `json:"netmask"`
	Network    string               `json:"network"`
	Primary    string               `json:"primary"`
	Interfaces map[string]Interface `json:"interfaces,omitempty"`
}

type Interface struct {
	IP        string    `json:"ip"`
	IP6       string    `json:"ip6"`
	MTU       int       `json:"mtu"`
	MAC       string    `json:"mac"`
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

var collectors []CollectorFunc

// Register adds a collector to the registry. It is typically called from init().
func Register(f CollectorFunc) {
	collectors = append(collectors, f)
}

// Collect runs all registered collectors and fills result. Returns the first error from any collector.
func Collect(result *Facts) error {
	for _, f := range collectors {
		if err := f(result); err != nil {
			return err
		}
	}
	return nil
}
