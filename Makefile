# Atlassian CLI Makefile

# Build variables
BINARY_NAME=atlassian-cli
VERSION?=$(shell git describe --tags --abbrev=0 2>/dev/null || echo "v1.0.0")
BUILD_DIR=bin
DIST_DIR=dist
COVERAGE_THRESHOLD=80

# Go variables
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=gofmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags "-X main.version=$(VERSION) -s -w"
BUILD_FLAGS=-trimpath

# Test flags
TEST_FLAGS=-v -race -timeout=30s
COVERAGE_FLAGS=-coverprofile=coverage.out -covermode=atomic

.PHONY: all build clean test coverage lint help install release dev-setup fmt vet security

all: clean fmt vet lint test build

## Development setup
dev-setup:
	@echo "Setting up development environment..."
	$(GOMOD) download
	$(GOMOD) tidy
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin v1.54.2; \
	fi
	@echo "✓ Development environment ready"

## Format code
fmt:
	@echo "Formatting code..."
	$(GOFMT) -s -w .
	$(GOCMD) mod tidy

## Vet code
vet:
	@echo "Vetting code..."
	$(GOVET) ./...

## Build the application
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "✓ Build completed: $(BUILD_DIR)/$(BINARY_NAME)"

## Build for all platforms
build-all: clean fmt vet
	@echo "Building for all platforms..."
	mkdir -p $(DIST_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(BUILD_FLAGS) $(LDFLAGS) -o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	@echo "✓ Multi-platform build completed"

## Run tests
test:
	@echo "Running tests..."
	$(GOTEST) $(TEST_FLAGS) ./...

## Run tests with race detection
test-race:
	@echo "Running tests with race detection..."
	$(GOTEST) $(TEST_FLAGS) -race ./...

## Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) $(TEST_FLAGS) $(COVERAGE_FLAGS) ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@$(GOCMD) tool cover -func=coverage.out | grep total | awk '{print "Total coverage: " $$3}'

## Check coverage threshold
coverage-check: coverage
	@COVERAGE=$$($(GOCMD) tool cover -func=coverage.out | grep total | awk '{print $$3}' | sed 's/%//'); \
	if [ "$$COVERAGE" -lt "$(COVERAGE_THRESHOLD)" ]; then \
		echo "❌ Coverage $$COVERAGE% is below threshold $(COVERAGE_THRESHOLD)%"; \
		exit 1; \
	else \
		echo "✓ Coverage $$COVERAGE% meets threshold $(COVERAGE_THRESHOLD)%"; \
	fi

## Run linter
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Run 'make dev-setup' first."; \
		exit 1; \
	fi

## Security scan
security:
	@echo "Running security scan..."
	@if command -v gosec >/dev/null 2>&1; then \
		gosec ./...; \
	else \
		echo "gosec not found. Installing..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
		gosec ./...; \
	fi

## Clean build artifacts
clean:
	@echo "Cleaning build artifacts..."
	$(GOCLEAN)
	rm -rf $(BUILD_DIR)
	rm -rf $(DIST_DIR)
	rm -f coverage.out coverage.html

## Install to system
install: build
	@echo "Installing $(BINARY_NAME)..."
	@if [ -w "/usr/local/bin" ]; then \
		cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
	else \
		sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/; \
	fi
	@echo "✓ Installed to /usr/local/bin/$(BINARY_NAME)"

## Install to GOPATH/bin
install-go: build
	@echo "Installing to GOPATH/bin..."
	cp $(BUILD_DIR)/$(BINARY_NAME) $(shell go env GOPATH)/bin/
	@echo "✓ Installed to $(shell go env GOPATH)/bin/$(BINARY_NAME)"

## Create release
release: 
	@echo "Creating release $(VERSION)..."
	./scripts/release.sh

## Tidy dependencies
tidy:
	@echo "Tidying dependencies..."
	$(GOMOD) tidy
	$(GOMOD) verify

## Update dependencies
update:
	@echo "Updating dependencies..."
	$(GOCMD) get -u ./...
	$(GOMOD) tidy

## Generate documentation
docs:
	@echo "Generating documentation..."
	@if [ -f "$(BUILD_DIR)/$(BINARY_NAME)" ]; then \
		mkdir -p docs/completion; \
		$(BUILD_DIR)/$(BINARY_NAME) completion bash > docs/completion/bash.sh; \
		$(BUILD_DIR)/$(BINARY_NAME) completion zsh > docs/completion/zsh.sh; \
		$(BUILD_DIR)/$(BINARY_NAME) completion fish > docs/completion/fish.sh; \
		echo "✓ Completion scripts generated"; \
	else \
		echo "Binary not found. Run 'make build' first."; \
	fi

## Run all quality checks
check: fmt vet lint security test coverage-check
	@echo "✓ All quality checks passed"

## Show help
help:
	@echo 'Atlassian CLI Build System'
	@echo ''
	@echo 'Usage:'
	@echo '  make <target>'
	@echo ''
	@echo 'Development:'
	@echo '  dev-setup        Set up development environment'
	@echo '  fmt              Format code'
	@echo '  vet              Vet code for issues'
	@echo '  lint             Run linter'
	@echo '  security         Run security scan'
	@echo '  check            Run all quality checks'
	@echo ''
	@echo 'Building:'
	@echo '  build            Build for current platform'
	@echo '  build-all        Build for all platforms'
	@echo '  clean            Clean build artifacts'
	@echo ''
	@echo 'Testing:'
	@echo '  test             Run tests'
	@echo '  test-race        Run tests with race detection'
	@echo '  coverage         Run tests with coverage'
	@echo '  coverage-check   Check coverage threshold'
	@echo ''
	@echo 'Installation:'
	@echo '  install          Install to /usr/local/bin'
	@echo '  install-go       Install to GOPATH/bin'
	@echo ''
	@echo 'Release:'
	@echo '  release          Create release with all platforms'
	@echo '  docs             Generate documentation'
	@echo ''
	@echo 'Maintenance:'
	@echo '  tidy             Tidy dependencies'
	@echo '  update           Update dependencies'
	@echo '  help             Show this help'