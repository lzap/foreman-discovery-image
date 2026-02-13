package facts

import (
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

func init() {
	Register(collectNetworking)
}

// defaultRouteInterface returns the name of the interface that has the default IPv4 route, or "" if none.
func defaultRouteInterface() string {
	f, err := os.Open("/proc/net/route")
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return ""
	}
	_ = scanner.Text() // skip header
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 3 {
			continue
		}
		iface, dest, flags := fields[0], fields[1], fields[3]
		if dest == "00000000" && (flags == "0003" || flags == "0002") {
			return iface
		}
	}
	return ""
}

// ipNetToNetmask returns dotted-decimal netmask for a *net.IPNet (e.g. "255.255.255.0").
func ipNetToNetmask(ipn *net.IPNet) string {
	if ipn == nil || len(ipn.Mask) != 4 {
		return ""
	}
	return net.IP(ipn.Mask).To4().String()
}

// ipNetToNetwork returns the network address of the IPNet as a string (e.g. "192.168.1.0" or "fe80::").
func ipNetToNetwork(ipn *net.IPNet) string {
	if ipn == nil {
		return ""
	}
	if ip := ipn.IP.To4(); ip != nil {
		return ipn.IP.Mask(ipn.Mask).To4().String()
	}
	return ipn.IP.Mask(ipn.Mask).String()
}

// ipNetToNetmaskPrefix returns the netmask as prefix length for IPv6 (e.g. "64"), or empty for IPv4.
func ipNetToNetmaskPrefix(ipn *net.IPNet) string {
	if ipn == nil || ipn.IP.To4() != nil {
		return ""
	}
	ones, _ := ipn.Mask.Size()
	return strconv.Itoa(ones)
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

		macStr := ""
		if len(iface.HardwareAddr) > 0 {
			macStr = iface.HardwareAddr.String()
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
			MAC:       macStr,
			Netmask:   firstNetmask,
			Network:   firstNetwork,
			Bindings:  bindings,
			Bindings6: bindings6,
		}

		if result.Custom == nil {
			result.Custom = make(map[string]any)
		}
		result.Custom["ipaddress_"+iface.Name] = firstIP
		result.Custom["macaddress_"+iface.Name] = macStr
		if firstIP6 != "" {
			result.Custom["ip6address_"+iface.Name] = firstIP6
		}

		if iface.Name == primaryName {
			result.Networking.Primary = iface.Name
			result.Networking.IP = firstIP
			result.Networking.MAC = macStr
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
