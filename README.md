# Vessel

**Vessel** is a tiny, educational container runtime written in Go. It demonstrates how containers work under the hood using Linux primitives—without relying on [containerd](https://containerd.io/), [runc](https://github.com/opencontainers/runc), or other large runtimes.

---

## Features

- **Namespaces:** Isolate global system resources (Mount, UTS, Network, IPC, PID).
- **Control Groups (cgroups):** Restrict CPU, memory, swap, and process usage.
- **Union File System:** OverlayFS for combining multiple image layers.
- **Image Management:** Pull and manage container images.
- **Networking:** Simple bridge and veth-based container networking.
- **CLI:** Docker-like commands for running and managing containers.

---

## Getting Started

### Prerequisites

- Linux (required for namespaces and cgroups)
- Go 1.15 or later

### Installation

```sh
go get -u github.com/0xc0d/vessel
```

## Usage

    Usage:
      vessel [command]
    
    Available Commands:
      exec        Run a command inside a existing Container.
      help        Help about any command
      images      List local images
      ps          List Containers
      run         Run a command inside a new Container.

## Examples

Run `/bin/sh` in `alpine:latest`

    vessel run alpine /bin/sh
    vessel run alpine # same as above due to alpine default command

Restart Nginx service inside a container with ID: 123456789123

    vessel exec 1234567879123 systemctrl restart nginx
    
List running containers

    vessel ps
    
List local images

    vessel images

## Notice
vessel, obviously, is not a production ready container manager tool.
