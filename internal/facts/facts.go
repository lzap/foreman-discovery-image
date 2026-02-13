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

	Custom map[string]any `json:"-"`
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
