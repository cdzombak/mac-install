SHELL:=/usr/bin/env bash

# nb. homebrew-releaser assumes the program name is == the repository name
BIN_NAME:=mac-install
BIN_VERSION:=$(shell ./.version.sh)

default: help
.PHONY: help  # via https://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help: ## Print help
	@grep -E '^[a-zA-Z_/-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: all
all: clean build-darwin-amd64 build-darwin-arm64 ## Build for macOS (amd64, arm64)

.PHONY: clean
clean: ## Remove build products (./out)
	rm -rf ./out

.PHONY: build
build: ## Build for the current platform & architecture to ./out
	mkdir -p out
	env CGO_ENABLED=0 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME} .

.PHONY: build-darwin-amd64
build-darwin-amd64: ## Build for macOS/amd64 to ./out
	env CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-darwin-amd64 .

.PHONY: build-darwin-arm64
build-darwin-arm64: ## Build for macOS/arm64 to ./out
	env CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -ldflags="-X main.version=${BIN_VERSION}" -o ./out/${BIN_NAME}-${BIN_VERSION}-darwin-arm64 .

.PHONY: test
test: ## Run the full test suite
	go test -v ./...

.PHONY: test-coverage
test-coverage: ## Run tests with coverage report
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

.PHONY: test-race
test-race: ## Run tests with race detection
	go test -race ./...

.PHONY: deps
deps: ## Download and tidy dependencies
	go mod download
	go mod tidy

.PHONY: lint
lint: ## Run golangci-lint
	golangci-lint run

.PHONY: fmt
fmt: ## Format code
	go fmt ./...

.PHONY: vet
vet: ## Vet code
	go vet ./...

.PHONY: run
run: build ## Build and run with example config
	./out/${BIN_NAME} -config install.example.yaml

.PHONY: install
install: ## Install binary to GOPATH/bin
	go install .

.PHONY: dev
dev: fmt vet test ## Development workflow (fmt, vet, test)

.PHONY: check
check: fmt vet lint test-race ## Full check (fmt, vet, lint, test-race)

.PHONY: validate-schema
validate-schema: ## Validate install.example.yaml against schema (requires ajv-cli)
	@if command -v ajv >/dev/null 2>&1; then \
		echo "Validating install.example.yaml against schema..."; \
		ajv validate -s schema.yaml -d install.example.yaml; \
	else \
		echo "ajv-cli not found. Install with: npm install -g ajv-cli"; \
	fi

.PHONY: validate-schema-syntax
validate-schema-syntax: ## Validate schema.yaml syntax (requires ajv-cli)
	@if command -v ajv >/dev/null 2>&1; then \
		echo "Validating schema.yaml syntax..."; \
		ajv compile -s schema.yaml; \
	else \
		echo "ajv-cli not found. Install with: npm install -g ajv-cli"; \
	fi
