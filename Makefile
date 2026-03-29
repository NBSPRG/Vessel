BINARY     := vessel
MODULE     := github.com/0xc0d/vessel
VERSION    := $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT     := $(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
BUILD_TIME := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ" 2>/dev/null || echo "unknown")

# Build flags: strip debug info + DWARF tables, embed version
LDFLAGS := -ldflags="-s -w \
	-X $(MODULE)/cmd.Version=$(VERSION) \
	-X $(MODULE)/cmd.Commit=$(COMMIT) \
	-X $(MODULE)/cmd.BuildTime=$(BUILD_TIME)"

# Disable CGO for a fully static binary (critical for multi-distro portability)
export CGO_ENABLED=0

.PHONY: all build build-upx size test test-unit test-distro \
        docker-build lint vet clean help

all: build

## build: Compile a static, stripped binary (~4-6 MB)
build:
	go build -trimpath $(LDFLAGS) -o $(BINARY) .
	@echo "Built $(BINARY) (CGO_ENABLED=0, trimpath, stripped)"

## build-upx: Compress binary with UPX (best-effort <2 MB; requires UPX in PATH)
build-upx: build
	@if command -v upx >/dev/null 2>&1; then \
		upx --best --lzma $(BINARY) && echo "UPX compression applied"; \
	else \
		echo "UPX not found — skipping compression. Install upx for sub-1MB binary."; \
	fi

## size: Report binary size after build
size: build
	@ls -lh $(BINARY) | awk '{printf "Binary size: %s\n", $$5}'

## vet: Run go vet static analysis
vet:
	go vet ./...

## lint: Alias for vet (extend with golangci-lint if available)
lint: vet
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found — ran go vet only"; \
	fi

## test: Run all unit tests
test: vet
	go test -race -count=1 ./... -v

## test-unit: Run only unit tests (no integration)
test-unit:
	go test -count=1 ./pkg/... -v

## test-distro: Validate across multiple Linux distros via Docker Compose
test-distro:
	docker compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker compose -f docker-compose.test.yml down --remove-orphans

## docker-build: Build the minimal production Docker image
docker-build:
	docker build -t $(BINARY):$(VERSION) -t $(BINARY):latest .
	@docker image ls $(BINARY)

## clean: Remove built artifacts
clean:
	rm -f $(BINARY)
	docker compose -f docker-compose.test.yml down --remove-orphans 2>/dev/null || true

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## /  /'
