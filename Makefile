.PHONY: all build test clean lint

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"

all: lint test build

build:
	@echo "Building..."
	go build $(LDFLAGS) -o bin/shelltask.exe cmd/shelltask/main.go

test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

clean:
	@echo "Cleaning..."
	rm -rf bin/ coverage.out

lint:
	@echo "Linting..."
	golangci-lint run

.PHONY: run
run:
	@echo "Running..."
	go run cmd/shelltask/main.go
