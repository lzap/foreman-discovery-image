package metric

import (
	"fmt"
	"log"
	"os/exec"

	"github.com/vishvananda/netlink"
	"golang.org/x/sys/unix"

	"fdi/internal/facts"
	"fdi/internal/opt"
)

// pxeRoutePriority is the route metric/priority for the PXE interface so it is
// preferred for outgoing traffic (lower = preferred). Default is often 100.
const pxeRoutePriority = 50

// EnsurePXE finds the PXE interface from discovery_bootif, and sets its route
// priority lower so that interface is preferred for outgoing communication. It
// does not do anything if just one interface is present, or when BOOTIF is not
// set.
func EnsurePXE() {
	var f facts.Facts
	if err := facts.Collect(&f, false, opt.CustomPath); err != nil {
		log.Printf("metric: failed to collect facts: %v", err)
		return
	}

	if len(f.Networking.Interfaces) < 2 {
		return
	}

	if f.DiscoveryBootIf.IsEmpty() {
		return
	}

	var pxeIfName string
	for name, iface := range f.Networking.Interfaces {
		if iface.MAC.Equal(f.DiscoveryBootIf) {
			pxeIfName = name
			break
		}
	}
	if pxeIfName == "" {
		return
	}

	link, err := netlink.LinkByName(pxeIfName)
	if err != nil {
		log.Printf("metric: link %s: %v", pxeIfName, err)
		return
	}

	filter := &netlink.Route{LinkIndex: link.Attrs().Index}
	for _, family := range []int{unix.AF_INET, unix.AF_INET6} {
		routes, err := netlink.RouteListFiltered(family, filter, netlink.RT_FILTER_OIF)
		if err != nil {
			log.Printf("metric: route list (family %d): %v", family, err)
			continue
		}
		for i := range routes {
			r := &routes[i]
			isDefault := r.Dst == nil || r.Dst.IP.IsUnspecified()
			if r.Priority == pxeRoutePriority {
				continue
			}

			r.Priority = pxeRoutePriority
			if opt.Debug {
				log.Printf("metric: route replace %s: %v", r.String(), r.Priority)
			}
			if err := netlink.RouteReplace(r); err != nil {
				log.Printf("metric: route replace %s: %v", r.String(), err)
				continue
			}

			if isDefault {
				go persistNMChange(pxeIfName, pxeRoutePriority)
			}
		}
	}
}

func persistNMChange(ifName string, priority int) {
	prio := fmt.Sprint(priority)
	if opt.Debug {
		log.Printf("metric: nmcli connection modify %s ip.route-metric %s", ifName, prio)
	}
	exec.Command("nmcli", "connection", "modify", ifName, "ipv4.route-metric", prio).Run()
	exec.Command("nmcli", "connection", "modify", ifName, "ipv6.route-metric", prio).Run()
}
