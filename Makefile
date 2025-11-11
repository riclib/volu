.PHONY: build test install clean help

# Build settings
BINARY_NAME=volu
BUILD_DIR=.
CMD_DIR=./cmd/volu

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

# Build flags
LDFLAGS=-ldflags="-s -w"

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

build: ## Build the binary
	$(GOBUILD) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Built $(BINARY_NAME)"

build-release: ## Build optimized release binary
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(CMD_DIR)
	@echo "Built optimized $(BINARY_NAME)"

test: ## Run tests
	$(GOTEST) -v -short ./...

test-all: ## Run all tests including integration tests
	$(GOTEST) -v ./...

test-coverage: ## Run tests with coverage
	$(GOTEST) -v -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

install: build ## Install the binary to /usr/local/bin
	sudo mv $(BINARY_NAME) /usr/local/bin/
	@echo "Installed $(BINARY_NAME) to /usr/local/bin/"

install-user: build ## Install the binary to ~/.local/bin
	mkdir -p ~/.local/bin
	mv $(BINARY_NAME) ~/.local/bin/
	@echo "Installed $(BINARY_NAME) to ~/.local/bin/"

clean: ## Clean build artifacts
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	@echo "Cleaned build artifacts"

deps: ## Download dependencies
	$(GOMOD) download
	$(GOMOD) tidy

fmt: ## Format code
	$(GOCMD) fmt ./...

vet: ## Run go vet
	$(GOCMD) vet ./...

lint: ## Run linters (requires golangci-lint)
	golangci-lint run

run: build ## Build and run with status command
	./$(BINARY_NAME) status

# Cross-compilation targets
build-linux-amd64: ## Build for Linux AMD64
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-amd64 $(CMD_DIR)

build-linux-arm64: ## Build for Linux ARM64
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm64 $(CMD_DIR)

build-linux-arm: ## Build for Linux ARM
	GOOS=linux GOARCH=arm $(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME)-linux-arm $(CMD_DIR)

build-all: build-linux-amd64 build-linux-arm64 build-linux-arm ## Build for all platforms
	@echo "Built binaries for all platforms"
