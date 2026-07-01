BINARY := lazy-openspec
PKG := ./cmd/lazy-openspec
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
LDFLAGS := -s -w -X main.version=$(VERSION)

.PHONY: build install run test vet fmt lint tidy clean

build: ## Build the binary into ./bin
	go build -ldflags "$(LDFLAGS)" -o bin/$(BINARY) $(PKG)

install: ## Install the binary onto your PATH
	go install -ldflags "$(LDFLAGS)" $(PKG)

run: ## Run against the current directory's openspec/ root
	go run $(PKG)

test: ## Run all tests
	go test ./...

vet: ## Run go vet
	go vet ./...

fmt: ## Format all Go files
	gofmt -w cmd internal

tidy: ## Tidy modules
	go mod tidy

clean: ## Remove build artifacts
	rm -rf bin
