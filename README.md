# Vessel

**Vessel** is a lightweight container runtime written in Go that demonstrates how containers work at the OS level using Linux primitives—without relying on [containerd](https://containerd.io/), [runc](https://github.com/opencontainers/runc), or other large runtimes.

---

## Features

- **Namespace isolation:** Mount, UTS, Network, IPC, and PID namespaces via Linux `clone()` and `setns()` syscalls.
- **cgroup resource management:** CPU quota, memory/swap limits, and PID limits via cgroup v1 controllers with **multi-tier allocation strategies** (see [Resource Tiers](#resource-tiers)).
- **Union filesystem:** OverlayFS union mounts across multiple image layers.
- **Image management:** Pull and cache OCI images from any container registry.
- **Container networking:** Bridge (`vessel0`) and virtual ethernet pair (veth) networking with automatic IP assignment.
- **Minimal binary:** Fully static binary (`CGO_ENABLED=0`) with stripped debug symbols.

---

## Build

### Requirements

- Linux (namespaces and cgroups are kernel features)
- Go 1.15+ (Go 1.21 recommended)
- GNU Make
- Docker + Docker Compose (for `make test-distro`)
- UPX (optional, for maximum binary compression)

### Quick build

```sh
make build          # static binary, stripped (~4-6 MB)
make build-upx      # UPX-compressed binary (~1-2 MB; requires upx in PATH)
make size           # build + report binary size
```

Binary optimisation flags applied:

| Flag | Effect |
|---|---|
| `CGO_ENABLED=0` | Fully static — no libc dependency |
| `-trimpath` | Removes local build paths from binary |
| `-ldflags="-s -w"` | Strips symbol table and DWARF debug info |
| UPX `--best --lzma` | LZMA compression (optional, best-effort <1 MB) |

### Docker build

```sh
make docker-build           # builds vessel:latest via multi-stage Dockerfile
docker run --rm vessel --help
```

---

## Resource Tiers

Vessel implements configurable **multi-tier resource allocation** via cgroup controllers, analogous to Kubernetes `LimitRange` / `ResourceQuota` tiers. Each tier defines a balanced CPU, memory, and process budget:

| Tier | Memory | Swap | CPUs | Max PIDs | Use case |
|---|---|---|---|---|---|
| `micro` | 128 MB | 64 MB | 0.10 | 50 | Scripts, utilities |
| `small` | 256 MB | 128 MB | 0.25 | 100 | Lightweight services |
| `medium` | 512 MB | 256 MB | 0.50 | 200 | Standard workloads |
| `large` | 1024 MB | 512 MB | 1.00 | 500 | Data-processing jobs |
| `xlarge` | 2048 MB | 1024 MB | 2.00 | 1000 | Memory-intensive workloads |

### Using tiers

```sh
# Run alpine with the medium resource tier
sudo vessel run --tier medium alpine /bin/sh

# Override a specific field — memory override wins over the tier
sudo vessel run --tier small --memory 512 alpine /bin/sh

# Traditional explicit limits (no tier)
sudo vessel run -m 256 -s 128 -c 0.5 -p 100 alpine /bin/sh
```

**Override semantics:** `--tier` sets the baseline; any explicit flag (`--memory`, `--swap`, `--cpus`, `--pids`) applied alongside it overrides only that field — matching Kubernetes LimitRange default-override behaviour.

---

## Usage

```
Usage:
  vessel [command]

Available Commands:
  exec        Run a command inside an existing container
  images      List local images
  ps          List running containers
  run         Run a command inside a new container

Run Flags (vessel run):
  -t, --tier string     Resource tier: micro|small|medium|large|xlarge
  -m, --memory int      Memory limit in MB (overrides --tier)
  -s, --swap int        Swap limit in MB (overrides --tier)
  -c, --cpus float      CPU limit (overrides --tier)
  -p, --pids int        Max processes (overrides --tier)
  -d, --detach          Run in background
      --host string     Container hostname
```

### Examples

```sh
# Run a shell in alpine with medium resource tier
sudo vessel run --tier medium alpine /bin/sh

# Run nginx detached with large tier
sudo vessel run --tier large -d nginx

# Exec into a running container
sudo vessel exec 1234567879123 /bin/sh

# List running containers
sudo vessel ps

# List cached images
sudo vessel images
```

---

## Testing

### Unit tests

```sh
make test           # go vet + go test ./... -race -v
make test-unit      # unit tests only (pkg/...)
```

### Cross-platform Linux distribution matrix

Validates build and unit tests across Ubuntu 20.04, Ubuntu 22.04, Debian Bookworm, and Alpine 3.18 using Docker Compose (no cloud account required):

```sh
make test-distro
# or directly:
docker compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

Each service installs Go 1.21, runs `go vet ./...` and `go test ./... -v`, and exits with a pass/fail code.

### Integration tests (requires Linux host)

Validates binary size, cgroup controller presence, and kernel namespace syscall support:

```sh
sudo bash scripts/test-integration.sh
```

Checks performed:
- Binary builds successfully (`CGO_ENABLED=0`)
- Binary size is within limits
- cgroup v1 controllers exist: `memory`, `cpu`, `pids`
- Kernel namespace support: `CONFIG_UTS_NS`, `CONFIG_IPC_NS`, `CONFIG_PID_NS`, `CONFIG_NET_NS`
- `/proc/self/exe` available for process re-execution

---

## Architecture

```
vessel run --tier medium alpine /bin/sh
        │
        ▼
  internal.Run()
  ├─ network.SetupBridge("vessel0")     # bridge + iptables
  ├─ ctr.SetupNetwork()                 # veth pair + netns
  ├─ image.Download() → OverlayFS mount # layer union mount
  └─ reexec(fork) with CLONE_NEWNS|CLONE_NEWUTS|CLONE_NEWIPC|CLONE_NEWPID
              │
              ▼
        internal.Fork()  [inside new namespaces]
        ├─ syscall.Sethostname()
        ├─ setns() → container netns
        ├─ cg.ApplyTier(medium)          # cgroup v1 limits
        │   ├─ memory.limit_in_bytes = 512 MB
        │   ├─ cpu.cfs_quota_us = 50000 (0.5 CPU)
        │   └─ pids.max = 200
        ├─ syscall.Chroot(rootfs)
        ├─ mount proc + sysfs
        └─ exec /bin/sh
```

**Storage paths:**

| Path | Purpose |
|---|---|
| `/var/lib/vessel/images/layers/` | Image layer tarballs |
| `/var/run/vessel/containers/<digest>/` | Container config + mnt |
| `/var/run/vessel/netns/<digest>` | Network namespace mount |
| `/sys/fs/cgroup/{memory,cpu,pids}/vessel/<digest>/` | cgroup hierarchy |

---

## Notice

Vessel is an educational container runtime demonstrating Linux kernel containerisation primitives. It is not intended for production workloads.
