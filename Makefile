

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get

# Main  package names
MAINDEV=./cmd/development/
MAINPROD=./cmd/production/

# Binary names
BINARY_NAME=api.exe

# Test package
TEST_PKG=./...

# Versioning
VERSION=$(shell git describe --tags --always --dirty="-dev")

# Setup linker flags option for build that interoperate with variable names in code
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAINDEV)

build-prod:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAINPROD)

test:
	$(GOTEST) $(TEST_PKG)

clean:
	$(GOCLEAN)

run:
	go run ./cmd/development/

run-prod:
	go run ./cmd/production/
	
lint:
	gofumpt -l -w .
	golangci-lint run  -v

all: lint test build

all-prod: lint test build-prod

.PHONY: all build test clean lint run