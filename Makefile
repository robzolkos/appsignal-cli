.PHONY: test test-unit test-e2e build clean tidy lint help

BINARY := $(CURDIR)/bin/appsignal-cli
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -X main.version=$(VERSION)

# Test configuration (set these or export as environment variables)
# export APPSIGNAL_TOKEN=your-token
# export APPSIGNAL_APP_ID=your-app-id

help:
	@echo "AppSignal CLI"
	@echo ""
	@echo "Usage:"
	@echo "  make build        Build the CLI"
	@echo "  make test-unit    Run unit tests (no API required)"
	@echo "  make test-e2e     Run e2e tests (requires API credentials)"
	@echo "  make test         Alias for test-unit"
	@echo "  make lint         Run golangci-lint"
	@echo "  make clean        Remove build artifacts"
	@echo "  make tidy         Tidy dependencies"
	@echo ""
	@echo "Environment variables (required for e2e tests):"
	@echo "  APPSIGNAL_TOKEN   API token"
	@echo "  APPSIGNAL_APP_ID  Application ID"

# Build CLI
build:
	@mkdir -p bin
	go build -ldflags "$(LDFLAGS)" -o $(BINARY) ./cmd/appsignal

# Run unit tests (no API required)
test-unit:
	go test -v ./internal/...

# Run e2e tests (requires API credentials)
test-e2e: build
	@if [ -z "$$APPSIGNAL_TOKEN" ]; then echo "Error: APPSIGNAL_TOKEN not set"; exit 1; fi
	@if [ -z "$$APPSIGNAL_APP_ID" ]; then echo "Error: APPSIGNAL_APP_ID not set"; exit 1; fi
	APPSIGNAL_TEST_BINARY=$(BINARY) go test -tags=e2e -v ./e2e/...

# Alias for test-unit
test: test-unit

# Lint
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -rf bin/
	go clean

# Tidy dependencies
tidy:
	go mod tidy

# Install locally
install: build
	cp $(BINARY) $(GOPATH)/bin/appsignal-cli
