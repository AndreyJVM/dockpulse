# Contributing to Dockpulse

Thank you for your interest in contributing to **Dockpulse**! We welcome all contributions: bug reports, documentation enhancements, feature proposals, and pull requests.

### Prerequisites

- **Go**: 1.27.3 or higher
- **Git**
- **Docker** (optional, recommended for testing Docker/Compose/Swarm integrations)
- **revive** linter (optional, install with `go install github.com/mgechev/revive@latest`)

### Getting Started

1. **Fork** the repository on GitHub:
   [https://github.com/AndreyJVM/dockpulse](https://github.com/AndreyJVM/dockpulse)

2. **Clone** your fork locally and configure the `upstream` remote:
   ```bash
   git clone git@github.com:<your-username>/dockpulse.git
   cd dockpulse
   git remote add upstream https://github.com/AndreyJVM/dockpulse.git
   git fetch upstream
   ```

3. **Create a branch** for your changes starting from `upstream/main`:
   ```bash
   git checkout -b feat/my-new-feature
   # or for bug fixes:
   git checkout -b fix/issue-description
   ```

### Project Architecture & Codebase Structure

Dockpulse is structured into clear and modular packages:

```text
dockpulse/
├── cmd/
│   └── dockpulse/          # Application entrypoint (main.go), CLI flags, and initialization
├── internal/               # Core internal packages (not importable by external projects)
│   ├── monitor/            # Monitoring engine and Docker engine integrations
│   │   ├── docker*.go      # Docker client, container stats, logs, inspect, exec, prune
│   │   ├── compose*.go     # Docker Compose service discovery and lifecycle
│   │   ├── swarm*.go       # Docker Swarm cluster nodes and services monitoring
│   │   └── system*.go      # Host system metrics (CPU, RAM, Disk, Network)
│   ├── tui/                # Terminal User Interface (Bubble Tea & Lip Gloss)
│   │   ├── views/          # Primary views (containers, Compose, networks, volumes, Swarm)
│   │   ├── modals/         # Interactive dialogs (logs, inspect, exec, prune, help)
│   │   ├── components/     # Reusable UI widgets (progress bar, sparkline, tables)
│   │   ├── theme/          # Color schemes, styles, and borders
│   │   └── update*.go      # Event handling, keybindings, and command dispatchers
│   └── i18n/               # English UI text catalog
├── docs/                   # Documentation and screenshots
└── debian/                 # Debian/Ubuntu packaging metadata
```

**Where to make your changes:**
- Adding or modifying metrics/Docker logic? Work in `internal/monitor/`.
- Adding a new dashboard view? Work in `internal/tui/views/`.
- Adding a popup modal or action? Work in `internal/tui/modals/`.
- Adding UI components or widgets? Work in `internal/tui/components/`.
- Updating UI text? Update `internal/i18n/dictionary.go`.

### Development & Testing

Use the provided `Makefile` targets to build, test, and verify your changes:

```bash
# Build standalone binary
make build

# Run unit tests
make test

# Run code style and quality checks (go vet + revive)
make lint

# Run locally in development mode
make run
```

### Commit Message Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification (see also `commit-template.txt`):

- `feat(...)`: New feature or capability
- `fix(...)`: Bug fix
- `docs(...)`: Documentation changes
- `refactor(...)`: Code refactoring without behavioral changes
- `test(...)`: Adding or updating tests
- `chore(...)`: Maintenance, dependency updates, version bumps

Example:
```text
feat(tui): add container memory limit indicator
docs: add contributing guide
fix(monitor): handle missing docker socket gracefully
```

### Submitting a Pull Request

1. Make sure all tests and linter checks pass (`make test && make lint`).
2. Push your branch to your fork:
   ```bash
   git push origin feat/my-new-feature
   ```
3. Open a Pull Request against the `main` branch of `AndreyJVM/dockpulse`.
4. Fill out the PR template with a clear description of the problem solved, what changes were made, and how they were tested.

### Automated Dependency Updates (Dependabot)

The repository uses **Dependabot** to keep dependencies and GitHub Actions workflows secure and up to date:
- Go module updates are grouped weekly.
- GitHub Actions updates are scheduled monthly.
- Automated CI workflows verify every dependency bump.
