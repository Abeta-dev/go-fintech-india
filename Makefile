.PHONY: all build test race bench lint fmt fmt-check coverage clean help

GOLANGCI_LINT := $(shell which golangci-lint 2>/dev/null || echo $(HOME)/go/bin/golangci-lint)

# Default target runs complete verification suite
all: fmt-check lint race coverage

# Build all packages
build:
	go build -v ./...

# Run standard unit tests
test:
	go test -v ./...

# Run unit tests with the Go race detector enabled
race:
	go test -race ./...

# Run high-throughput benchmarks with memory allocation metrics
bench:
	go test -run=^$$ -bench=. -benchmem ./...

# Run golangci-lint across all packages
lint:
	$(GOLANGCI_LINT) run ./...

# Auto-format all Go source files with gofmt
fmt:
	gofmt -w .

# Enforce that all Go source files are formatted with gofmt
fmt-check:
	@test -z "$$(gofmt -l .)" || (echo "❌ Unformatted files detected. Run 'make fmt':" && gofmt -l . && exit 1)
	@echo "✅ All Go files are formatted with gofmt."

# Run comprehensive statement coverage gate script (enforces >= 90%)
coverage:
	./scripts/check_coverage.sh

# Clean build and test artifacts
clean:
	rm -f coverage.out coverage.html coverage.json *.test
