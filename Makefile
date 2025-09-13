.PHONY: build test lint fmt generate install deps tag

export GOBIN ?= $(shell pwd)/bin

BINARY_NAME=contexture
BINARY_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
MAIN_PACKAGE=./cmd/contexture
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-X main.Version=$(VERSION)

MOCKERY = $(GOBIN)/mockery
GOIMPORTS = $(GOBIN)/goimports
GOFUMPT = $(GOBIN)/gofumpt
GOLANGCI_LINT = $(GOBIN)/golangci-lint

$(MOCKERY):
	go install github.com/vektra/mockery/v3@latest

$(GOIMPORTS):
	go install golang.org/x/tools/cmd/goimports@latest

$(GOFUMPT):
	go install mvdan.cc/gofumpt@latest

$(GOLANGCI_LINT):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(GOBIN) v2.4.0
 
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test: build
	@go test -race -cover -coverprofile=coverage.out -timeout=5m ./...

lint: $(GOLANGCI_LINT)
	@echo "Running linter..."
	$(GOLANGCI_LINT) run

fmt: $(GOIMPORTS) $(GOFUMPT)
	@echo "Formatting code..."
	@$(GOIMPORTS) -w $(GO_FILES)
	@$(GOFUMPT) -e -l -w $(GO_FILES)

generate: $(MOCKERY)
	@echo "Generating mocks..."
	$(MOCKERY)

install:
	@echo "Installing $(BINARY_NAME) $(VERSION)..."
	@go install -ldflags "$(LDFLAGS)" $(MAIN_PACKAGE)

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy

tag:
	@if [ -z "$(VERSION)" ]; then \
		echo "Usage: make tag VERSION=v1.0.0"; \
		exit 1; \
	fi
	@echo "Creating and pushing tag $(VERSION)..."
	@git tag -a $(VERSION) -m "Release $(VERSION)"
	@git push origin $(VERSION)
	@echo "Tag $(VERSION) created and pushed. GitHub Actions will now create the release."

build-release:
	@echo "Building $(BINARY_NAME) $(VERSION) for release..."
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags "$(LDFLAGS) -s -w" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)