FROM quay.io/centos-bootc/centos-bootc:stream10

# Configuration for nm to set the default route priority for the PXE interface
COPY pre-nm-pxe-setup.sh /usr/local/bin/pre-nm-pxe-setup.sh
RUN chmod +x /usr/local/bin/pre-nm-pxe-setup.sh

# Configuration to run the script before NetworkManager starts
RUN mkdir -p /etc/systemd/system/NetworkManager.service.d/
COPY pre-nm-setup.conf /etc/systemd/system/NetworkManager.service.d/pre-nm-setup.conf
