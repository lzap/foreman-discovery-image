package facts

import (
	"bytes"
	"fdi/internal/cmdline"
	"log"
	"net"
)

var zeroMAC = net.HardwareAddr{0, 0, 0, 0, 0, 0}

func init() {
	Register(discoveryBootIf)
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
