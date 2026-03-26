.PHONY: build install all clean plugin-install-global plugin-install-local

BINARY_ARM64 := build/gsd-watch-darwin-arm64
BINARY_AMD64 := build/gsd-watch-darwin-amd64
INSTALL_DIR  := $(HOME)/.local/bin
LDFLAGS      := -ldflags="-s -w"
CMD_SRC      := ./cmd/gsd-watch/
CODESIGN_ID  := Apple Development

build: $(BINARY_ARM64) $(BINARY_AMD64)

$(BINARY_ARM64):
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $@ $(CMD_SRC)
	codesign --force --sign "$(CODESIGN_ID)" $@

$(BINARY_AMD64):
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $@ $(CMD_SRC)
	codesign --force --sign "$(CODESIGN_ID)" $@

install: build
	mkdir -p $(INSTALL_DIR)
	@ARCH=$$(uname -m); \
	if [ "$$ARCH" = "arm64" ]; then \
		cp $(BINARY_ARM64) $(INSTALL_DIR)/gsd-watch; \
	else \
		cp $(BINARY_AMD64) $(INSTALL_DIR)/gsd-watch; \
	fi
	@echo "Installed to $(INSTALL_DIR)/gsd-watch"

all: install

clean:
	rm -rf build/

plugin-install-global:
	cp commands/gsd-watch.md $(HOME)/.claude/commands/gsd-watch.md
	@echo "Installed slash command to $(HOME)/.claude/commands/gsd-watch.md"

plugin-install-local:
	mkdir -p .claude/commands
	cp commands/gsd-watch.md .claude/commands/gsd-watch.md
	@echo "Installed slash command to .claude/commands/gsd-watch.md"
