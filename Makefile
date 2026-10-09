VERSION ?= 0.1.0
BINARY_NAME = vc-env
BUILD_DIR = build
LDFLAGS = -ldflags "-X main.Version=$(VERSION)"

.PHONY: build build-all test test-docker clean install docs-build docs-serve

## build: Build for current platform
build:
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/vc-env

## build-all: Cross-compile for all supported platforms
build-all:
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 ./cmd/vc-env
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 ./cmd/vc-env
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./cmd/vc-env
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./cmd/vc-env

## test: Run unit tests
test:
	go test -v -race ./...

## test-docker: Run Docker integration tests
test-docker:
	./scripts/test-docker.sh

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## install: Install to GOPATH/bin
install:
	go install $(LDFLAGS) ./cmd/vc-env

## help: Show this help
help:
	@echo "Available targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'

.venv-docs: docs/requirements.txt
	python3 -m venv .venv-docs
	.venv-docs/bin/pip install -q -r docs/requirements.txt
	touch .venv-docs

## docs-build: Build the MkDocs site (strict) into site/
docs-build: .venv-docs
	.venv-docs/bin/mkdocs build --strict

## docs-serve: Serve the MkDocs site with live reload at http://127.0.0.1:8000/vc-env/
docs-serve: .venv-docs
	.venv-docs/bin/mkdocs serve
