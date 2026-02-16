IMAGE_NAME := fdi-image
MOCK_CMDLINE_FILE := /tmp/fdi_mock_cmdline

BUILD_ARGS :=
ifdef PASSWORD
    BUILD_ARGS += --build-arg PASSWORD=$(PASSWORD)
endif
OUTPUT := ./output
VM_PATH := $(OUTPUT)
VM_BRIDGE := virbr0
TEST_MAC := "52:54:00:21:13:15"
QEMU_URL := "qemu:///system"
TAR_NAME := bootc-centos-10-pxe-tar-xz-x86_64

.PHONY: all run clean
GOSRC := $(shell find cmd internal -type f) go.mod go.sum
CSRC := Containerfile *.sh *.conf *.service
VMSRC := *.toml
all: .container

fdi: $(SRC)
	go build -o fdi ./cmd

.container: fdi $(CSRC)
	sudo podman build $(BUILD_ARGS) -t $(IMAGE_NAME) . -f Containerfile
	touch .container

run-container: .container
	@if [ -z "$(CMDLINE)" ]; then \
		echo "Error: Please provide CMDLINE='your boot args'"; \
		exit 1; \
	fi
	@echo "BOOTIF=01-$(shell echo $(TEST_MAC) | tr ':' '-') $(CMDLINE)" > $(MOCK_CMDLINE_FILE)
	sudo podman run --rm -it --privileged \
		--device /dev/net/tun:/dev/net/tun \
		--network bridge:mac=$(TEST_MAC) \
		-v $(MOCK_CMDLINE_FILE):/proc/cmdline:ro \
		localhost/$(IMAGE_NAME):latest

.vm: .container $(VMSRC)
	@test -d $(OUTPUT) || mkdir $(OUTPUT)
	sudo podman run -it --rm --privileged \
		--pull=newer --security-opt label=type:unconfined_t \
		-v ./config.toml:/config.toml:ro \
		-v /var/lib/containers/storage:/var/lib/containers/storage \
		-v $(OUTPUT):/output \
		ghcr.io/osbuild/image-builder-cli:latest \
		build --blueprint /config.toml --bootc-default-fs ext4 --with-metrics \
		--bootc-ref localhost/$(IMAGE_NAME) pxe-tar-xz
	touch .vm

extract:
	sudo bash -c ' \
		set -e; \
		pushd $(OUTPUT)/$(TAR_NAME) > /dev/null; \
		tar xJf $(TAR_NAME).tar.xz; \
		echo rootfs.img | cpio -H newc --quiet -L -o > rootfs.cpio; \
		cat initrd.img rootfs.cpio > combined.img; \
		chmod 644 combined.img vmlinuz; \
		rm rootfs.cpio; \
		popd > /dev/null; \
	'

#run-vm: .vm
run-vm:
	virt-install --connect "$(QEMU_URL)" --transient --destroy-on-exit \
		--name "$(IMAGE_NAME)" --vcpus 2 --memory 4096 \
		--import --os-variant centos-stream10 \
		--boot kernel=$(VM_PATH)/$(TAR_NAME)/vmlinuz,initrd=$(VM_PATH)/$(TAR_NAME)/combined.img,kernel_args="root=live:/rootfs.img rd.live.image rw console=tty0 console=ttyS0 ostree=/ostree/boot.1/default/0b9a28494bca0cadcd14e6aa0d0e4af37b956795754983d990d616bf9d2538c0/0" \
		--network bridge=$(VM_BRIDGE),mac=$(TEST_MAC) \
		--network bridge=$(VM_BRIDGE)

clean:
	rm -f fdi .container
	rm -f $(MOCK_CMDLINE_FILE)
	podman rmi $(IMAGE_NAME) || true
