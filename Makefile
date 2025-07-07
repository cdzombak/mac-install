# macOS Automated Setup System Makefile

BINARY_NAME=mac-install
MAIN_PATH=.
BUILD_DIR=build

.PHONY: build test clean install run help

# Default target
all: build

# Build the binary
build:
	go build -o $(BINARY_NAME) $(MAIN_PATH)

# Build for release with optimizations
build-release:
	CGO_ENABLED=0 go build -ldflags="-w -s" -o $(BINARY_NAME) $(MAIN_PATH)

# Run tests
test:
	go test ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run tests with race detection
test-race:
	go test -race ./...

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	go clean

# Install dependencies
deps:
	go mod download
	go mod tidy

# Lint the code (requires golangci-lint)
lint:
	golangci-lint run

# Format code
fmt:
	go fmt ./...

# Vet code
vet:
	go vet ./...

# Run the program with example config
run: build
	./$(BINARY_NAME) -config install.example.yaml

# Install the binary to $GOPATH/bin
install:
	go install $(MAIN_PATH)

# Development workflow: format, vet, test
dev: fmt vet test

# Full check: format, vet, lint, test with race detection
check: fmt vet lint test-race

# Validate YAML files against schema (requires ajv-cli: npm install -g ajv-cli)
validate-schema:
	@if command -v ajv >/dev/null 2>&1; then \
		echo "Validating install.example.yaml against schema..."; \
		ajv validate -s schema.yaml -d install.example.yaml; \
	else \
		echo "ajv-cli not found. Install with: npm install -g ajv-cli"; \
	fi

# Validate schema syntax
validate-schema-syntax:
	@if command -v ajv >/dev/null 2>&1; then \
		echo "Validating schema.yaml syntax..."; \
		ajv compile -s schema.yaml; \
	else \
		echo "ajv-cli not found. Install with: npm install -g ajv-cli"; \
	fi

# Help target
help:
	@echo "Available targets:"
	@echo "  build               - Build the binary"
	@echo "  build-release       - Build optimized release binary"
	@echo "  test                - Run tests"
	@echo "  test-coverage       - Run tests with coverage report"
	@echo "  test-race           - Run tests with race detection"
	@echo "  clean               - Clean build artifacts"
	@echo "  deps                - Download and tidy dependencies"
	@echo "  lint                - Run linter (requires golangci-lint)"
	@echo "  fmt                 - Format code"
	@echo "  vet                 - Vet code"
	@echo "  run                 - Build and run with example config"
	@echo "  install             - Install binary to GOPATH/bin"
	@echo "  dev                 - Development workflow (fmt, vet, test)"
	@echo "  check               - Full check (fmt, vet, lint, test-race)"
	@echo "  validate-schema     - Validate install.example.yaml against schema (requires ajv-cli)"
	@echo "  validate-schema-syntax - Validate schema.yaml syntax (requires ajv-cli)"
	@echo "  help                - Show this help message"