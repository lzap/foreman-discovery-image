package facts

import (
	"bufio"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// readTrimmedFile reads path and returns its contents trimmed of surrounding whitespace.
// Returns empty string on error or missing file.
func readTrimmedFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// mapArch maps a Go architecture to a conventional name (e.g. amd64 -> x86_64).
func mapArch(goarch string) string {
	switch goarch {
	case "amd64":
		return "x86_64"
	default:
		return goarch
	}
}

var pciIDsPaths = []string{
	"/usr/share/hwdata/pci.ids",
	"/usr/share/pci.ids",
	"/usr/share/misc/pci.ids",
}

// resolvePCIVendor returns a human-readable name for a PCI vendor ID (e.g. "0x1af4" -> "Red Hat, Inc.").
// If the ID is not hex or not found in pci.ids, returns the original string or "0x" + id.
func resolvePCIVendor(vendor string) string {
	vendor = strings.TrimSpace(vendor)
	if vendor == "" {
		return ""
	}
	vendor = strings.TrimPrefix(strings.ToLower(vendor), "0x")
	if len(vendor) != 4 {
		return vendor
	}
	for _, r := range vendor {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') {
			continue
		}
		return vendor
	}
	for _, path := range pciIDsPaths {
		if name := lookupPCIVendor(path, vendor); name != "" {
			return name
		}
	}
	return "0x" + vendor
}

func lookupPCIVendor(path, vendorID string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if len(line) < 5 {
			continue
		}
		if line[0] == '#' || line[0] == '\t' || line[0] == 'C' {
			continue
		}
		if strings.EqualFold(line[:4], vendorID) {
			name := strings.TrimSpace(line[4:])
			if name != "" {
				return name
			}
			return ""
		}
	}
	return ""
}

// readBlockSize reads size in 512-byte sectors from path and returns total bytes, or -1 on error.
func readBlockSize(path string) int64 {
	s := readTrimmedFile(path)
	if s == "" {
		return -1
	}
	sectors, err := strconv.ParseInt(s, 10, 64)
	if err != nil || sectors < 0 {
		return -1
	}
	return sectors * 512
}

// pxelinuxToMAC converts a PXELinux-style BOOTIF/fdi.pxmac value to colon-separated MAC string.
// Format: "01-52-54-00-94-9e-52" or "01-52:54:00:94:9e:52" -> "52:54:00:94:9e:52".
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
