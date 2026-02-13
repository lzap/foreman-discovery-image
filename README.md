# Foreman Discovery Image

This branch contains a modernized version of FDI based on bootable containers.
The whole workflow and services is refactored to a single Go-based service that
performs all the logic. The base system was changed from CentOS 9 to 10 with as
little customizations as possible.

The aim is backward compatibility with the original FDI and the following kernel
command line options are fully functional:

* proxy.url - URL to Foreman or Proxy (if omitted DNS SRV lookup is made)
* proxy.type - endpoint type: “foreman” or “proxy”
* fdi.uploadsleep - seconds between facter runs (30 by default)
* fdi.dns_nameserver - nameserver to use for DNS SRV record
* fdi.dns_search - search domain to use for DNS SRV record
* fdi.dns_ndots - ndots option to use for DNS SRV record

Features that are not carried over:

* fdi.ssh - configure ssh daemon after start (1 to enable)
* fdi.rootpw - configure ssh daemon password (plain string)
* fdi.countdown - initial countdown in seconds before registration attempt (45 by default)
* fdi.cachefacts - number of fact uploads without caching (0 by default)
* fdi.zips - extensions to download
* fdi.zipserver - override TFTP server reported by DHCP
* fdi.initnet=all/bootif - initialize all or pxe NICs (default) during startup
* fdi.pxmac - alternative to BOOTIF (just use BOOTIF instead)
* fdi.px* - PXE-less workflow (described below)
* fdi.dhcp_timeout - DHCP NetworkManager timeout in seconds (300 by default)
* fdi.ipwait - wait time for IP to be available in proxy SSL cert start (120 by default)
* fdi.nmwait - nmcli –wait option for NetworkManager (120 by default)
* fdi.proxy_cert_days - number of days HTTPS self-signed cert is valid (999 by default)
* fdi.script - base64 encoded boot script
* fdi.vlan.primary - VLAN ID to set for primary interface

Setting root account using kernel command line was insecure, this is now
possible via podman variable. When it is set, password and ssh is configured
accordingly.

Caching of facts will not be necessary as facts will not be implemented via
Facter anymore.

ZIP extensions and scripts are replaced by easy customization via bootable
containers.

Various workarounds or settings for network initialization will be dropped,
because FDI will make full use of NetworkManager and network metrics to properly
setup default route automatically. Countdown will not be necessary and PXE-less
workflow will be completely removed as it was never fully supported by Red Hat.

VLAN is not implemented, customers need to customize FDI in order to achieve this.

We need to reach out to customers who are using the unsupported PXE-less
workflow for virtualized environments and discuss possible options.

## Usage

The fdi binary is an application spawned by systemd as a service which performs
all actions.

### Command-line options

* `-facts` — only print collected facts to stdout (JSON)
* `-debug` — safe mode (no reboots or configuration actions) and print debugging information to stderr (e.g. top 5 slowest collectors)
* `-once` — perform a single facts upload to the endpoint and exit
* `-url` — endpoint base URL (e.g. `https://foreman.example.com`). If omitted, the kernel command line option `proxy.url` is used
* `-type` — endpoint type: `foreman` or `proxy` (default: `foreman`). If omitted, the kernel command line option `proxy.type` is used
* `-custom_path` — path to search for `fdi-fact-*` executables in addition to PWD (default: `/usr/local/bin`)
* `-uploadsleep` — seconds to wait between fact uploads in service mode. If 0 (default), the kernel command line option `fdi.uploadsleep` is used, or 30 if unset.

When started without `-facts` or `-once`, **service mode** runs: a loop that collects facts, uploads them to the endpoint, then sleeps for the configured interval. Shutdown is graceful on SIGINT (Ctrl-C) or SIGTERM. Endpoint URL and type are taken from `proxy.url` and `proxy.type` on the kernel command line when not set by flags.

## Custom facts

Any executables in the current directory (PWD) or in the path given by `-custom_path` (default `/usr/local/bin`), whose names start with `fdi-fact-`, will be executed. Output is trimmed from
whitespace and added to the fact ouput. All stderr output is returned in
`_stderr` and exit code that is not zero is returned in `_code`. Example:

```
cat ./fdi-fact-abc
#!/bin/sh
echo Hi
```

Make sure it is executable and:

```
fdi -facts
{
  "abc": "Hi"
}
```


## Building

Requirements:

* Go 1.26+
* podman

To build the container:

  make

The root account is fully locked, to unlock it, set a password and also start ssh:

  make PASSWORD=redhat

## Testing in a container

To test the image in podman:

  make run PASSWORD=redhat CMDLINE="proxy.url=https://192.168.99.97"

## Testing in a VM

  virt-install --name fdi --memory 3300 --vcpus 2 --nodisks --os-variant centos-stream10 \
    --network network=default,model=virtio,mac=52:54:00:12:34:56 \
    --network network=default,model=virtio \
    --boot kernel=/var/lib/libvirt/images/tar/vmlinuz,initrd=/var/lib/libvirt/images/tar/combined.img,kernel_args="root=live:/rootfs.img rd.live.image rw console=tty0 console=ttyS0 fdi.ssh=1 proxy.url=https://XXX BOOTIF=01-52-54-00-12-34-56 ostree=/ostree/boot.1/default/XXX/0" \
    --serial pty --console pty,target_type=serial --graphics none

## TODO

* Test multiple interfaces in a VM (podman cannot work due to hardcoded routes)
* Test via qcow2
* Test via PXE

The original code can be found in `legacy/` folder.
