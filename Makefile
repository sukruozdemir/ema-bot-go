# Binary names
BINARY_NAME=ema-bot
CONFIG_TOOL=ema-bot-config
WINDOWS_BINARY=$(BINARY_NAME).exe
LINUX_BINARY=$(BINARY_NAME)-linux
MACOS_BINARY=$(BINARY_NAME)-macos
MACOS_ARM_BINARY=$(BINARY_NAME)-macos-arm64

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build flags
LDFLAGS=-ldflags="-s -w"

# Directories
BUILD_DIR=./build
CMD_APP=./cmd/app
CMD_CONFIG=./cmd/config

.PHONY: all build clean test coverage fmt vet lint run run-config help install deps cross-compile

## all: Build all binaries
all: clean deps build

## build: Build the application binaries
build:
	@echo "Building application..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_APP)
	$(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(CONFIG_TOOL) $(CMD_CONFIG)
	@echo "Build complete! Binaries in $(BUILD_DIR)/"

## build-dev: Build without optimization (for debugging)
build-dev:
	@echo "Building application (development mode)..."
	@mkdir -p $(BUILD_DIR)
	$(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME) $(CMD_APP)
	$(GOBUILD) -o $(BUILD_DIR)/$(CONFIG_TOOL) $(CMD_CONFIG)
	@echo "Development build complete!"

## run: Run the application directly
run:
	@echo "Running application..."
	$(GOCMD) run $(CMD_APP)/main.go

## run-config: Run the configuration tool
run-config:
	@echo "Running configuration tool..."
	$(GOCMD) run $(CMD_CONFIG)/main.go

## test: Run all tests
test:
	@echo "Running tests..."
	$(GOTEST) -v -race -timeout 30s ./...

## test-short: Run short tests only
test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

## coverage: Run tests with coverage
coverage:
	@echo "Running tests with coverage..."
	$(GOTEST) -v -race -coverprofile=coverage.out -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	$(GOTEST) -bench=. -benchmem ./...

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

## vet: Run go vet
vet:
	@echo "Running go vet..."
	$(GOVET) ./...

## lint: Run all linting tools
lint: fmt vet
	@echo "Linting complete!"

## deps: Download dependencies
deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

## clean: Clean build artifacts
clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	@rm -rf $(BUILD_DIR)
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"

## install: Install the binaries to GOPATH/bin
install:
	@echo "Installing binaries..."
	$(GOCMD) install $(CMD_APP)
	$(GOCMD) install $(CMD_CONFIG)
	@echo "Installation complete!"

## cross-compile: Build for all platforms
cross-compile: clean
	@echo "Cross-compiling for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	@echo "Building for Windows (AMD64)..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(WINDOWS_BINARY) $(CMD_APP)
	
	@echo "Building for Linux (AMD64)..."
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(LINUX_BINARY) $(CMD_APP)
	
	@echo "Building for macOS (Intel)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(MACOS_BINARY) $(CMD_APP)
	
	@echo "Building for macOS (Apple Silicon)..."
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BUILD_DIR)/$(MACOS_ARM_BINARY) $(CMD_APP)
	
	@echo "Cross-compilation complete! Binaries in $(BUILD_DIR)/"

## race: Run with race detector
race:
	@echo "Running with race detector..."
	$(GOCMD) run -race $(CMD_APP)/main.go

## profile-cpu: Run with CPU profiling
profile-cpu:
	@echo "Running with CPU profiling..."
	$(GOCMD) run -cpuprofile=cpu.prof $(CMD_APP)/main.go
	@echo "CPU profile saved to cpu.prof"
	@echo "Analyze with: go tool pprof cpu.prof"

## profile-mem: Run with memory profiling
profile-mem:
	@echo "Running with memory profiling..."
	$(GOCMD) run -memprofile=mem.prof $(CMD_APP)/main.go
	@echo "Memory profile saved to mem.prof"
	@echo "Analyze with: go tool pprof mem.prof"

## check: Run all checks (fmt, vet, test)
check: fmt vet test
	@echo "All checks passed!"

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'
