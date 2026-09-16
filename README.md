# NEXUS — Developer Command Center

Terminal-based Developer Command Center written in Go.
Single terminal interface to inspect the dev machine, diagnose problems,
check Docker, Git, networking, and optionally ask an LLM for help.

Primary platform: macOS Apple Silicon (kept portable to Linux where practical).

## Architecture

```
nexus/
  main.go            # entrypoint -> cmd.Execute()
  cmd/               # Cobra commands (root + 8 subcommands)
  config/            # env-var configuration (no secrets hardcoded)
  internal/logger/   # stderr slog logger (stdlib only)
  internal/          # shared helpers (future phases)
  ui/                # Bubble Tea / Lip Gloss UI (future phase, empty stub)
  docs/              # per-phase reports
```

- Small packages, explicit errors, no global mutable state (except cobra's `--verbose` flag binding, which is idiomatic).
- Commands work independently; optional deps (Docker, Git, AI) never crash the binary.
- Never runs destructive commands automatically. AI is optional.

## Installation

Requires Go 1.22+.

```bash
git clone <repo-url> nexus && cd nexus
go mod download
go build -o nexus .
```

## Usage

```bash
./nexus --help
./nexus dashboard
./nexus system
./nexus processes
./nexus docker
./nexus git
./nexus network
./nexus doctor
./nexus ask "why is my build slow?"
./nexus --verbose system   # debug logging to stderr
```

`nexus` with no args shows help.

### Configuration (environment variables)

| Variable | Purpose | Default |
|---|---|---|
| `NEXUS_DEBUG` | Verbose debug logging (`1`/`true`) | off |
| `NEXUS_NO_COLOR` / `NO_COLOR` | Disable color (respected in later UI phases) | off |
| `NEXUS_AI_MODEL` | Model name for `nexus ask` | provider default |
| `NEXUS_OPENAI_API_KEY` | Enables `nexus ask`; empty = AI disabled | empty |

No config files, no hardcoded secrets. API keys are never printed.

## Development commands

```bash
make build   # go build -o nexus .
make run     # go run . --help
make test    # go test ./...
make fmt     # gofmt -w + go vet
make clean   # rm -f nexus
```

## Current feature status (Phase 5 — Git)

| Command | Status |
|---|---|
| `nexus dashboard` | Stub — prints "not implemented yet" |
| `nexus system` | Done — OS/arch/hostname, CPU, memory, disk, uptime via gopsutil |
| `nexus processes` | Done — top-N table (PID/CPU/MEM/MEMORY/NAME/STATUS), `--limit`, `--sort cpu\|memory`, read-only |
| `nexus docker` | Done — read-only `docker ps` table (CONTAINER/IMAGE/STATUS/PORTS/CREATED), `--all`, graceful unavailable message |
| `nexus git` | Done — read-only overview (branch, clean/dirty, modified/staged/untracked, latest commit+date), `--short`, graceful outside repos |
| `nexus network` | Stub |
| `nexus doctor` | Stub |
| `nexus ask` | Stub (optional, env-configured later) |

Foundation done: Go module, `cmd/`, `config/`, `internal/`, `ui/` structure,
env config, stderr logging, Makefile, tests. No monitoring/Docker/Git/network/AI logic yet.
