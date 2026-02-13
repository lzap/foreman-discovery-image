package facts

import (
	"fdi/internal/cmdline"
	"log"
)

func init() {
	Register("discovery_bootif", discoveryBootIf)
}

func discoveryBootIf(result *Facts) error {
	macs := systemMACs()
	if len(macs) == 0 {
		result.DiscoveryBootIf = EmptyMAC
		log.Printf("discovery_bootif: no network interfaces found, using %s", EmptyMAC.String())
		return nil
	}
	if len(macs) == 1 {
		result.DiscoveryBootIf = MACFromHardwareAddr(macs[0])
		return nil
	}

	raw := cmdline.Get("BOOTIF")
	if raw == "" {
		result.DiscoveryBootIf = EmptyMAC
		log.Printf("discovery_bootif: multiple interfaces but no BOOTIF on kernel command line, using %s", EmptyMAC.String())
		return nil
	}

	bootMAC := ParseBOOTIF(raw)
	if bootMAC.IsEmpty() {
		result.DiscoveryBootIf = EmptyMAC
		log.Printf("discovery_bootif: invalid BOOTIF value %q, using %s", raw, EmptyMAC.String())
		return nil
	}

	for _, hw := range macs {
		if MACFromHardwareAddr(hw).Equal(bootMAC) {
			result.DiscoveryBootIf = MACFromHardwareAddr(hw)
			return nil
		}
	}

	var macStrs []string
	for _, hw := range macs {
		macStrs = append(macStrs, MACFromHardwareAddr(hw).String())
	}
	result.DiscoveryBootIf = EmptyMAC
	log.Printf("discovery_bootif: option %q does not match any interface (have %v), using %s", raw, macStrs, EmptyMAC.String())
	return nil
}
