.PHONY: all build test clean lint

VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X main.Version=$(VERSION)"
SHELLTASK_SRCS := $(wildcard cmd/shelltask/*.go)
all: lint test build

build:
	@echo "Building..."
	go build $(LDFLAGS) -o bin/shelltask.exe $(SHELLTASK_SRCS)

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
	go run $(SHELLTASK_SRCS)
