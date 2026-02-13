#!/bin/bash

# Extract the MAC from BOOTIF (format 01-aa-bb-cc...)
BOOTIF_RAW=$(grep -oP 'BOOTIF=\K[0-9a-fA-F:-]+' /proc/cmdline)

if [ -z "$BOOTIF_RAW" ]; then
    echo "No BOOTIF found. Exiting."
    exit 0
fi

# Clean MAC (remove 01- prefix, change - to :, lowercase)
BOOTIF_MAC=$(echo "$BOOTIF_RAW" | sed 's/^01-//;s/-/:/g' | tr 'A-F' 'a-f')

# Ensure the default metric is 100 and no other connections are auto-defaulted
cat <<EOF > "/etc/NetworkManager/conf.d/99-default-metrics.conf"
[main]
no-auto-default=*

[connection]
ipv4.route-metric=100
ipv6.route-metric=100
EOF

# Find the interface name associated with this MAC (compare lowercase)
IFACE_NAME=""
for dev in /sys/class/net/*; do
    if [ -f "$dev/address" ]; then
        dev_mac=$(tr 'A-F' 'a-f' < "$dev/address")
        if [ "$dev_mac" = "$BOOTIF_MAC" ]; then
            IFACE_NAME=$(basename "$dev")
            break
        fi
    fi
done

if [ -z "$IFACE_NAME" ]; then
    echo "Could not map MAC $BOOTIF_MAC to an interface."
    exit 1
fi

# Write the PXE-specific connection (this will be the default route)
mkdir -p /etc/NetworkManager/system-connections/
cat <<EOF > "/etc/NetworkManager/system-connections/${IFACE_NAME}.nmconnection"
[connection]
id=${IFACE_NAME}
interface-name=${IFACE_NAME}
type=ethernet
autoconnect=true

[ipv4]
method=auto
route-metric=50

[ipv6]
method=auto
route-metric=50
EOF

chmod 600 "/etc/NetworkManager/system-connections/${IFACE_NAME}.nmconnection"
echo "Configured $IFACE_NAME as primary PXE interface."
