package facts

import (
	"log"
	"net"
	"os"
	"strings"
)

func init() {
	Register("networking", collectNetworking)
}

func collectNetworking(result *Facts) error {
	primaryName := defaultRouteInterface()
	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "localhost"
	}

	result.Networking = Networking{
		Hostname:   hostname,
		FQDN:       hostname,
		Interfaces: make(map[string]Interface),
	}

	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("networking: failed to list interfaces: %v", err)
		return nil
	}

	var ifaceNames []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		var macAddr MAC
		if len(iface.HardwareAddr) == 6 || len(iface.HardwareAddr) == 20 {
			macAddr = MACFromHardwareAddr(iface.HardwareAddr)
		}

		var bindings, bindings6 []Binding
		var firstIP, firstIP6, firstNetmask, firstNetwork string

		for _, a := range addrs {
			ipn, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			if ipn.IP.To4() != nil {
				ipStr := ipn.IP.To4().String()
				netmask := ipNetToNetmask(ipn)
				network := ipNetToNetwork(ipn)
				bindings = append(bindings, Binding{
					Address: ipStr,
					Netmask: netmask,
					Network: network,
				})
				if firstIP == "" {
					firstIP, firstNetmask, firstNetwork = ipStr, netmask, network
				}
			} else {
				ipStr := ipn.IP.String()
				netmask := ipNetToNetmaskPrefix(ipn)
				network := ipNetToNetwork(ipn)
				bindings6 = append(bindings6, Binding{
					Address: ipStr,
					Netmask: netmask,
					Network: network,
				})
				if firstIP6 == "" {
					firstIP6 = ipStr
				}
			}
		}

		if firstIP == "" && firstIP6 == "" {
			continue
		}

		ifaceNames = append(ifaceNames, iface.Name)
		result.Networking.Interfaces[iface.Name] = Interface{
			IP:        firstIP,
			IP6:       firstIP6,
			MTU:       iface.MTU,
			MAC:       macAddr,
			Netmask:   firstNetmask,
			Network:   firstNetwork,
			Bindings:  bindings,
			Bindings6: bindings6,
		}

		if result.Custom == nil {
			result.Custom = make(map[string]any)
		}
		result.Custom["ipaddress_"+iface.Name] = firstIP
		result.Custom["macaddress_"+iface.Name] = macAddr
		if firstIP6 != "" {
			result.Custom["ip6address_"+iface.Name] = firstIP6
		}

		if iface.Name == primaryName {
			result.Networking.Primary = iface.Name
			result.Networking.IP = firstIP
			result.Networking.MAC = macAddr
			result.Networking.Netmask = firstNetmask
			result.Networking.Network = firstNetwork
		}
	}

	result.Interfaces = strings.Join(ifaceNames, ",")

	// If no default route but exactly one interface, treat it as primary.
	if result.Networking.Primary == "" && len(result.Networking.Interfaces) == 1 {
		for name, iface := range result.Networking.Interfaces {
			result.Networking.Primary = name
			result.Networking.IP = iface.IP
			result.Networking.MAC = iface.MAC
			result.Networking.Netmask = iface.Netmask
			result.Networking.Network = iface.Network
			break
		}
	}

	return nil
}
