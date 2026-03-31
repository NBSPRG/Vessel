# Vessel — Claude Code Context

## Project
Lightweight container runtime in Go. Uses Linux kernel primitives (namespaces, cgroups, OverlayFS).
Module: `github.com/0xc0d/vessel`

## Language & Build
- Go 1.21+ (module: go 1.15 minimum)
- CGO_ENABLED=0 — fully static binary, no libc dependency
- Build: `make build` or `go build -trimpath -ldflags="-s -w" -o vessel .`
- Binary name: `vessel`

## Code Style
- Standard Go formatting (`gofmt`)
- `go vet ./...` must pass before commits
- golangci-lint if available
- No external frameworks — keep it minimal

## Testing
- Unit tests: `make test-unit` → `go test -count=1 ./pkg/... -v`
- Full tests: `make test` → `go test -race -count=1 ./... -v`
- Distro tests: `make test-distro` (requires Docker, runs across Linux distros)
- WARNING: namespace/cgroup tests require Linux root privileges — they will fail on macOS or Windows

## Key Packages
- `cmd/` — CLI entry points (cobra)
- `internal/` — core runtime (namespaces, cgroups, networking)
- `pkg/` — reusable packages (unit-testable without root)

## Security Rules
- No hardcoded credentials or tokens
- All secrets via environment variables
- vessel requires root at runtime (Linux only) — this is intentional
- SAST: run `gosec ./...` and `govulncheck ./...`

## Docker
- Multi-stage build: golang:1.21-alpine → alpine:3.18
- Runtime deps: iptables, iproute2 (already in Dockerfile)
- Image: chandannbsprg/vessel
- Build: `make docker-build`

## Deployment Notes
- vessel is a CLI/runtime tool, not a web service
- No Kubernetes deployment needed for vessel itself
- Releases: static binary published to GitHub Releases + Docker Hub
