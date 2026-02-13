# Variables
IMAGE_NAME := fdi-image
TEST_NET := fdi-net
MOCK_CMDLINE_FILE := /tmp/fdi_mock_cmdline

# Conditional Build Argument logic
BUILD_ARGS :=
ifdef PASSWORD
    BUILD_ARGS += --build-arg PASSWORD=$(PASSWORD)
endif

SOURCES := Containerfile pre-nm-setup.sh pre-nm-setup.conf fdi.service

# Targets
.PHONY: all fdi run clean

all: .container

# Build the Go binary
fdi:
	go build -o fdi ./cmd

# Build the container image
.container: fdi $(SOURCES)
	podman build $(BUILD_ARGS) -t $(IMAGE_NAME) . -f Containerfile
	touch .container

# Run the container with a mocked kernel command line
run: .container
	@if [ -z "$(CMDLINE)" ]; then \
		echo "Error: Please provide CMDLINE='your boot args'"; \
		exit 1; \
	fi
	$(eval TEST_MAC := 00:11:22:33:44:55)
	@podman network exists $(TEST_NET) || podman network create $(TEST_NET)
	@echo "BOOTIF=01-$(shell echo $(TEST_MAC) | tr ':' '-') $(CMDLINE)" > $(MOCK_CMDLINE_FILE)
	podman run --rm -it \
		--cap-add=NET_ADMIN \
		--cap-add=SYS_ADMIN \
		--security-opt seccomp=unconfined \
		--device /dev/net/tun:/dev/net/tun \
		--network bridge \
		--network $(TEST_NET):mac=$(TEST_MAC) \
		-v $(MOCK_CMDLINE_FILE):/proc/cmdline:ro \
		$(IMAGE_NAME):latest

clean:
	rm -f fdi .container
	rm -f $(MOCK_CMDLINE_FILE)
	podman rmi $(IMAGE_NAME) || true
