<div align="center">

# Dockpulse

[![CI](https://github.com/AndreyJVM/dockpulse/actions/workflows/ci.yml/badge.svg)](https://github.com/AndreyJVM/dockpulse/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/AndreyJVM/dockpulse?color=blue&logo=github)](https://github.com/AndreyJVM/dockpulse/releases)
[![Go Version](https://img.shields.io/badge/Go-1.27.3+-00ADD8?logo=go)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux-lightgrey)](https://github.com/AndreyJVM/dockpulse/releases)

**A fast, lightweight terminal TUI for Docker containers and resources.**<br>

[Contributing](CONTRIBUTING.md)

---

</div>

### Key Features

- **Docker Containers**:
  - Live CPU% and memory stats with dynamic color coding.
  - Interactive sorting by CPU%, RAM, Name, or Status (`s`, `o`).
  - Live Network I/O (RX/TX) and Block I/O (Read/Write) metrics in details panel and inspect modal.
  - Healthcheck statuses (`[healthy]`, `[unhealthy]`, `[starting]`).
  - Search & filter by name, image, ID, or health status (`/`).
  - Logs viewer with scroll support (`L`).
  - Detailed container inspection modal (`Enter` / `I`).
  - Interactive shell session (`E` — `/bin/bash` or `/bin/sh`) inside container.
  - Power control: Start / Stop (`S`), Restart (`P`), Force Kill (`K` — `SIGKILL`).
  - Safe container removal with confirmation (`x`, `d`).
  - Docker System Prune modal (`C`) with disk reclaim estimates and confirmation.
  - Socket support: Docker daemon, Rootless Docker, and Podman (<2 ms latency).
- **Docker Compose**: Automatic project/service grouping, aggregated stack health, live stack logs (`L`), service control (`U` / `S` / `P`), and stack teardown (`D` / down) with confirmation.
- **Docker Networks**: Network driver, scope, connected container count, subnet, gateway, and Compose ownership.
- **Docker Volumes**: Volume driver, scope, reference count, disk usage, mountpoint, and Compose ownership.
- **Docker Swarm**: Cluster node roles (Leader, Manager, Worker), service listings, replica counts with degradation highlights, and live service logs (`L`).
- **English Interface**: Consistent English UI, CLI help, diagnostics, and documentation.
- **Zero Dependencies**: Standalone binary (<12 MB) with zero external runtime dependencies. Runs on Linux.

### Installation

#### Pre-built `.deb` Package (Debian / Ubuntu / Linux Mint)

```bash
# AMD64 (x86_64)
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.9.0_amd64.deb
sudo dpkg -i dockpulse_0.9.0_amd64.deb

# ARM64
wget https://github.com/AndreyJVM/dockpulse/releases/latest/download/dockpulse_0.9.0_arm64.deb
sudo dpkg -i dockpulse_0.9.0_arm64.deb
```

#### Pre-built Binaries (Linux)

Download the latest archive from [Releases](https://github.com/AndreyJVM/dockpulse/releases):

```bash
# Linux AMD64
tar -xzf dockpulse_v0.9.0_linux_amd64.tar.gz
./dockpulse
```

#### Build from Source (Go 1.27.3+)

```bash
git clone https://github.com/AndreyJVM/dockpulse.git
cd dockpulse
make build
./dockpulse
```

### Command-line Options

```text
Usage:
  dockpulse [options]

Options:
  -i, --interval <duration>  Auto-refresh interval (default: 2s, e.g., 1s, 500ms)
  -t, --tab <1-5|name>       Initial active tab (1: Docker, 2: Compose, 3: Networks, 4: Volumes, 5: Swarm) (default: 1)
  -v, --version              Show version information
  -h, --help                 Show help message
```

### Hotkeys

| Key | Action |
|---|---|
| `1`..`5` / `Tab` / `←`/`→` | Switch tabs (Docker, Compose, Networks, Volumes, Swarm) |
| `r`, `R` | Force immediate refresh |
| `↑` / `↓` (or `k` / `j`) | Navigate rows / select items |
| `/` | Search & filter (`Esc` to clear) |
| `s`, `S`, `o`, `O` | Sort containers (CPU, Memory, Name, Status) |
| `L` | View logs (scroll with `↑`/`↓`/`PgUp`/`PgDn`) |
| `Enter` / `I` | Inspect container (details, mounts, network, ports, health) |
| `E` | Attach interactive shell (`/bin/bash` or `/bin/sh`) |
| `S` | Start / Stop container or Compose stack |
| `P` | Restart container or Compose stack |
| `K` | Force kill (`SIGKILL`) container |
| `x`, `d` | Remove container (with confirmation) |
| `U` | Start Compose stack |
| `D` | Compose down (`docker compose down`) with confirmation |
| `C` | Docker system prune (cleanup unused resources) |
| `?` | Toggle help / hotkey shortcuts modal |
| `q` / `Esc` / `Ctrl+C` | Quit application |

### Contributing

Contributions are welcome! Please refer to [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing guidelines, and pull request workflow.

---

## Screenshots

> Place real UI screenshots in [`docs/screenshots/`](docs/screenshots/).
