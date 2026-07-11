# Get version information from git and system
VERSION ?= (unknown)
GIT_COMMIT := $(shell git rev-parse --short HEAD)
BUILD_DATE := $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')

# Package and binary name
BINARY := prometheus-mcp-server
PACKAGE := $(shell go list -m)

# Build flags
LDFLAGS := -s -w
LDFLAGS += -X $(PACKAGE)/internal/buildinfo.Number=$(VERSION)
LDFLAGS += -X $(PACKAGE)/internal/buildinfo.GitCommit=$(GIT_COMMIT)
LDFLAGS += -X $(PACKAGE)/internal/buildinfo.BuildDate=$(BUILD_DATE)

# Default target
all: build

# Build the binary
build:
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) main.go

# Install the binary
install:
	go install -ldflags "$(LDFLAGS)" .

# Clean build artifacts
clean:
	rm -f $(BINARY)

# Help target
help:
	@echo "Available targets:"
	@echo "  all     - Build the binary (default)"
	@echo "  build   - Build the binary"
	@echo "  install - Install the binary"
	@echo "  clean   - Remove build artifacts"
	@echo "  help    - Show this help"

.PHONY: all build install clean help
