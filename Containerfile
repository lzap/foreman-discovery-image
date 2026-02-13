FROM quay.io/centos-bootc/centos-bootc:stream10

# Install dracut/squashfs for live image and all supporting packages
RUN dnf -y install \
    dracut-live \
    squashfs-tools \
    NetworkManager \
    openssh-clients \
    openssh-server \
    passwd \
    openssl \
    rsync \
    ethtool

# Mask unnecessary units which do not work in containers
RUN systemctl mask nfs-client.target rpcbind.service rpcbind.socket

# Configuration for nm to set the default route priority for the PXE interface
COPY pre-nm-setup.sh /usr/local/bin/pre-nm-setup.sh
RUN chmod +x /usr/local/bin/pre-nm-setup.sh

# Configuration to run the script before NetworkManager starts
RUN mkdir -p /etc/systemd/system/NetworkManager.service.d/
COPY pre-nm-setup.conf /etc/systemd/system/NetworkManager.service.d/pre-nm-setup.conf

# Service file for fdi and fdi itself
COPY fdi /usr/local/bin/fdi
COPY fdi.service /etc/systemd/system/fdi.service
RUN systemctl enable fdi.service

ARG PASSWORD
RUN if [ -n "$PASSWORD" ]; then \
        echo "root:${PASSWORD}" | chpasswd && \
        usermod -U root && \
        systemctl enable sshd.service && \
        sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config; \
    fi
