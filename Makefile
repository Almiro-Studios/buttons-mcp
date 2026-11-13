VERSION ?= $(shell git describe --tags --always 2>/dev/null || echo dev)

build:
	go build -ldflags "-X main.version=$(VERSION)" -o bin/buttons-mcp ./cmd/buttons-mcp

install:
	go install -ldflags "-X main.version=$(VERSION)" ./cmd/buttons-mcp

test:
	go vet ./... && go test ./...

.PHONY: build install test
