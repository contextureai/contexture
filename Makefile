.PHONY: build test lint fmt generate integration e2e docker-integration docker-e2e install deps tag

BINARY_NAME=contexture
BINARY_DIR=bin
GO_FILES=$(shell find . -name '*.go' -not -path "./vendor/*")
MAIN_PACKAGE=./cmd/contexture
VERSION?=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT?=$(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE?=$(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS=-X main.Version=$(VERSION)

$(mockery):
	go install github.com/vektra/mockery/v3@latest

$(goimports):
	go install golang.org/x/tools/cmd/goimports@latest

$(gofumpt):
	go install mvdan.cc/gofumpt@latest

$(golangci-lint):
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.4.0
 
build:
	@echo "Building $(BINARY_NAME) $(VERSION)..."
	@mkdir -p $(BINARY_DIR)
	@go build -ldflags "$(LDFLAGS)" -o $(BINARY_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)

test:
	@go test -race -cover -coverprofile=coverage.out -timeout=5m $(shell go list ./... | grep -v -E '/(e2e|integration)$$')

lint: $(golangci-lint)
	@echo "Running linter..."
	@golangci-lint run

fmt: $(goimports) $(gofumpt)
	@echo "Formatting code..."
	@goimports -w $(GO_FILES)
	@gofumpt -e -l -w $(GO_FILES)

generate: $(mockery)
	@echo "Generating mocks..."
	@mockery

integration:
	@go test -v -race -timeout=10m ./integration/...

e2e:
	@go test -race -timeout=15m ./e2e/...

docker-integration:
	@docker build -f build/Dockerfile.test -t contexture-test . > /dev/null 2>&1 || echo "Using existing image..."
	@docker run --rm \
		-v "$(PWD)":/app \
		-v contexture-go-mod-cache:/go/pkg/mod \
		-v contexture-go-cache:/go/cache \
		-w /app \
		contexture-test sh -c "make integration"

docker-e2e:
	@docker build -f build/Dockerfile.test -t contexture-test . > /dev/null 2>&1 || echo "Using existing image..."
	@mkdir -p test-output && chmod 755 test-output
	@docker run --rm \
		-v "$(PWD)/e2e":/app/e2e \
		-v "$(PWD)/test-output":/app/test-output \
		-v contexture-go-mod-cache:/go/pkg/mod \
		-v contexture-go-cache:/go/cache \
		-w /app \
		-e CONTEXTURE_DOCKER_TEST=true \
		contexture-test sh -c "make e2e"

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
