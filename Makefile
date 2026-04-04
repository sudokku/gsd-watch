.PHONY: build-darwin build-linux build-all install clean plugin-install-global plugin-install-local

BINARY_ARM64       := build/gsd-watch-darwin-arm64
BINARY_AMD64       := build/gsd-watch-darwin-amd64
BINARY_LINUX_ARM64 := build/gsd-watch-linux-arm64
BINARY_LINUX_AMD64 := build/gsd-watch-linux-amd64
INSTALL_DIR        := $(HOME)/.local/bin
LDFLAGS            := -ldflags="-s -w"
CMD_SRC            := ./cmd/gsd-watch/
CODESIGN_ID        := Apple Development

build-darwin: $(BINARY_ARM64) $(BINARY_AMD64)

build-linux: $(BINARY_LINUX_ARM64) $(BINARY_LINUX_AMD64)

build-all: build-darwin build-linux

$(BINARY_ARM64):
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $@ $(CMD_SRC)
	codesign --force --sign "$(CODESIGN_ID)" --options runtime --timestamp $@

$(BINARY_AMD64):
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $@ $(CMD_SRC)
	codesign --force --sign "$(CODESIGN_ID)" --options runtime --timestamp $@

$(BINARY_LINUX_ARM64):
	mkdir -p build/
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $@ $(CMD_SRC)

$(BINARY_LINUX_AMD64):
	mkdir -p build/
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $@ $(CMD_SRC)

install:
	mkdir -p $(INSTALL_DIR)
	@OS=$$(uname -s); \
	ARCH=$$(uname -m); \
	if [ "$$ARCH" = "aarch64" ]; then ARCH=arm64; fi; \
	if [ "$$ARCH" = "x86_64" ]; then ARCH=amd64; fi; \
	if [ "$$OS" = "Darwin" ] && [ "$$ARCH" = "arm64" ]; then \
		BINARY=$(BINARY_ARM64); \
	elif [ "$$OS" = "Darwin" ] && [ "$$ARCH" = "amd64" ]; then \
		BINARY=$(BINARY_AMD64); \
	elif [ "$$OS" = "Linux" ] && [ "$$ARCH" = "arm64" ]; then \
		BINARY=$(BINARY_LINUX_ARM64); \
	elif [ "$$OS" = "Linux" ] && [ "$$ARCH" = "amd64" ]; then \
		BINARY=$(BINARY_LINUX_AMD64); \
	else \
		echo "Unsupported platform: $$OS/$$ARCH"; exit 1; \
	fi; \
	if [ ! -f "$$BINARY" ]; then \
		echo "Binary not found: $$BINARY"; \
		echo "Run 'make build-darwin' or 'make build-linux' first"; \
		exit 1; \
	fi; \
	cp "$$BINARY" $(INSTALL_DIR)/gsd-watch; \
	echo "Installed $$BINARY to $(INSTALL_DIR)/gsd-watch"

clean:
	rm -rf build/

plugin-install-global:
	cp commands/gsd-watch.md $(HOME)/.claude/commands/gsd-watch.md
	@echo "Installed slash command to $(HOME)/.claude/commands/gsd-watch.md"

plugin-install-local:
	mkdir -p .claude/commands
	cp commands/gsd-watch.md .claude/commands/gsd-watch.md
	@echo "Installed slash command to .claude/commands/gsd-watch.md"
