package facts

import (
	"bytes"
	"fdi/internal/cmdline"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

var zeroMAC = net.HardwareAddr{0, 0, 0, 0, 0, 0}

func init() {
	Register(discoveryBootIf)
}

// pxelinuxToMAC converts a PXELinux-style BOOTIF/fdi.pxmac value to colon-separated MAC string.
// Format: "01-52-54-00-94-9e-52" or "01-52:54:00:94:9e:52" -> "52:54:00:94:9e:52".
// Strips leading ARP type (e.g. 01-), then replaces any dashes with colons.
func pxelinuxToMAC(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	_, after, ok := strings.Cut(s, "-")
	if !ok {
		return ""
	}
	return strings.ToLower(strings.ReplaceAll(after, "-", ":"))
}

// systemMACs reads all non-loopback interface MACs from /sys/class/net.
// Returns nil on error or if no interfaces found.
func systemMACs() []net.HardwareAddr {
	entries, err := os.ReadDir("/sys/class/net")
	if err != nil {
		log.Printf("discovery_bootif: cannot read /sys/class/net: %v", err)
		return nil
	}
	var macs []net.HardwareAddr
	for _, e := range entries {
		if e.Name() == "lo" {
			continue
		}
		data, err := os.ReadFile(filepath.Join("/sys/class/net", e.Name(), "address"))
		if err != nil {
			continue
		}
		hw, err := net.ParseMAC(strings.TrimSpace(string(data)))
		if err != nil {
			continue
		}
		macs = append(macs, hw)
	}
	return macs
}

func discoveryBootIf(result *Facts) error {
	macs := systemMACs()
	if len(macs) == 0 {
		result.DiscoveryBootIf = zeroMAC.String()
		log.Printf("discovery_bootif: no network interfaces found, using %s", zeroMAC.String())
		return nil
	}
	if len(macs) == 1 {
		result.DiscoveryBootIf = macs[0].String()
		return nil
	}

	raw := cmdline.Get("BOOTIF")
	if raw == "" {
		raw = cmdline.Get("fdi.pxmac")
	}
	if raw == "" {
		result.DiscoveryBootIf = zeroMAC.String()
		log.Printf("discovery_bootif: multiple interfaces but no BOOTIF or fdi.pxmac on kernel command line, using %s", zeroMAC.String())
		return nil
	}

	colonMAC := pxelinuxToMAC(raw)
	if colonMAC == "" {
		result.DiscoveryBootIf = zeroMAC.String()
		log.Printf("discovery_bootif: invalid BOOTIF/fdi.pxmac value %q, using %s", raw, zeroMAC.String())
		return nil
	}
	bootHW, err := net.ParseMAC(colonMAC)
	if err != nil {
		result.DiscoveryBootIf = zeroMAC.String()
		log.Printf("discovery_bootif: invalid BOOTIF/fdi.pxmac value %q, using %s", raw, zeroMAC.String())
		return nil
	}

	for _, hw := range macs {
		if bytes.Equal(hw, bootHW) {
			result.DiscoveryBootIf = hw.String()
			return nil
		}
	}

	var macStrs []string
	for _, hw := range macs {
		macStrs = append(macStrs, hw.String())
	}
	result.DiscoveryBootIf = zeroMAC.String()
	log.Printf("discovery_bootif: option %q does not match any interface (have %v), using %s", raw, macStrs, zeroMAC.String())
	return nil
}
