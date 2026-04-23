BINARY    := kamusm-go
BUILD_DIR := bin
VERSION   := $(shell grep 'const Version' kamusm-zd/version.go | cut -d'"' -f2)
LDFLAGS   := -s -w

.PHONY: build vet clean install run test help

## build: Compile binary to bin/
build:
	go build -ldflags "$(LDFLAGS)" -o $(BUILD_DIR)/$(BINARY) .

## vet: Run static analysis
vet:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf $(BUILD_DIR)

## install: Install binary to GOPATH/bin
install:
	go install -ldflags "$(LDFLAGS)" .

## run: Build and run with ARGS (e.g. make run ARGS="bakiye")
run: build
	./$(BUILD_DIR)/$(BINARY) $(ARGS)

## test: Run build + vet
test: build vet

## version: Show current version
version:
	@echo $(VERSION)

## help: Show available targets
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
