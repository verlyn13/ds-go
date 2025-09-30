.PHONY: build install test clean run lint release help

# This Makefile provides backward compatibility while primarily using mise tasks
# Run 'mise tasks' to see all available tasks with descriptions

# Variables
BINARY_NAME=ds
MAIN_PATH=./cmd/ds
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

# Default target
help:
	@echo "ds-go Development Commands"
	@echo ""
	@echo "This project uses mise for task management. Available commands:"
	@echo ""
	@echo "  mise run build      - Build the binary"
	@echo "  mise run install    - Build and install to /usr/local/bin"
	@echo "  mise run test       - Run tests with coverage"
	@echo "  mise run lint       - Run golangci-lint"
	@echo "  mise run fmt        - Format code"
	@echo "  mise run dev        - Run in development mode"
	@echo "  mise run clean      - Clean build artifacts"
	@echo "  mise run ci         - Run CI pipeline locally"
	@echo "  mise run validate   - Validate system compliance"
	@echo ""
	@echo "Legacy Makefile targets are preserved for compatibility."
	@echo "Run 'mise tasks' for complete list with descriptions."

# Build the binary - delegates to mise
build:
	@mise run build

# Install locally - delegates to mise
install:
	@mise run install

# Run without installing
run:
	go run ${MAIN_PATH} $(ARGS)

# Run tests - delegates to mise
test:
	@mise run test

# Run benchmarks - delegates to mise
bench:
	@mise run bench

# Lint with golangci-lint - delegates to mise
lint:
	@mise run lint

# Format code - delegates to mise
fmt:
	@mise run fmt

# Clean build artifacts - delegates to mise
clean:
	@mise run clean

# Create a release with goreleaser - delegates to mise
release:
	@mise run release

# Create a snapshot release - delegates to mise
snapshot:
	@mise run snapshot

# Update dependencies - delegates to mise
deps:
	@mise run deps

# Run with race detector - delegates to mise
race:
	@mise run race $(ARGS)

# Profile CPU - delegates to mise
profile-cpu:
	@mise run profile-cpu $(ARGS)

# Profile memory - delegates to mise
profile-mem:
	@mise run profile-mem $(ARGS)

# Quick status check - delegates to mise
status:
	@mise run status

# Quick fetch - delegates to mise
fetch:
	@mise run fetch

# Run CI pipeline locally
ci:
	@mise run ci

# Validate system compliance
validate:
	@mise run validate