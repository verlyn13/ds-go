.PHONY: build install test clean run lint release

# Variables
BINARY_NAME=ds
MAIN_PATH=./cmd/ds
VERSION=$(shell git describe --tags --always --dirty)
LDFLAGS=-ldflags "-s -w -X main.version=${VERSION}"

# Build the binary
build:
	go build ${LDFLAGS} -o ${BINARY_NAME} ${MAIN_PATH}

# Install locally
install: build
	sudo mv ${BINARY_NAME} /usr/local/bin/

# Run without installing
run:
	go run ${MAIN_PATH} $(ARGS)

# Run tests
test:
	go test -v -race -cover ./...

# Run benchmarks
bench:
	go test -bench=. -benchmem ./...

# Lint with golangci-lint
lint:
	golangci-lint run ./...

# Format code
fmt:
	go fmt ./...
	gofumpt -w .

# Clean build artifacts
clean:
	rm -f ${BINARY_NAME}
	rm -rf dist/

# Create a release with goreleaser
release:
	goreleaser release --clean

# Create a snapshot release (for testing)
snapshot:
	goreleaser release --snapshot --clean

# Update dependencies
deps:
	go get -u ./...
	go mod tidy

# Run with race detector
race:
	go run -race ${MAIN_PATH} $(ARGS)

# Profile CPU
profile-cpu:
	go run -cpuprofile=cpu.prof ${MAIN_PATH} $(ARGS)
	go tool pprof cpu.prof

# Profile memory
profile-mem:
	go run -memprofile=mem.prof ${MAIN_PATH} $(ARGS)
	go tool pprof mem.prof

# Quick status check
status: build
	./$(BINARY_NAME) status

# Quick fetch
fetch: build
	./$(BINARY_NAME) fetch