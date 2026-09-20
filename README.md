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
  internal/system/   # OS/CPU/memory/disk/uptime via gopsutil
  internal/process/  # read-only process table
  internal/docker/   # read-only docker ps via CLI
  internal/git/      # read-only repo overview via CLI
  internal/network/  # stdlib diagnostics (DNS, dial, interfaces)
  internal/doctor/   # PASS/WARN/FAIL checks reusing the above
  internal/ai/       # optional OpenAI-compatible assistant
  internal/logger/   # stderr slog logger (stdlib only)
  ui/                # Bubble Tea / Lip Gloss dashboard (screens, loaders, styles)
  docs/              # per-phase reports
```

- Small packages, explicit errors, no global mutable state (except cobra's `--verbose` flag binding, which is idiomatic).
- Commands work independently; optional deps (Docker, Git, AI) never crash the binary.
- Never runs destructive commands automatically. AI is optional.

## Installation

Requires Go 1.24+.

```bash
git clone <repo-url> nexus && cd nexus
go mod download
go build -o nexus .
```

## Usage

```bash
./nexus              # interactive dashboard (needs a TTY)
./nexus dashboard    # same dashboard
./nexus --help
./nexus system
./nexus processes
./nexus docker
./nexus git
./nexus network
./nexus doctor
./nexus ask "why is my build slow?"
./nexus --verbose system   # debug logging to stderr
```

Keys in the dashboard: `↑`/`↓` navigate · `Enter` open · `r` refresh · `Esc` home · `q` quit.

### Configuration (environment variables)

| Variable | Purpose | Default |
|---|---|---|
| `NEXUS_DEBUG` | Verbose debug logging (`1`/`true`) | off |
| `NEXUS_NO_COLOR` / `NO_COLOR` | Disable color (dashboard uses plain styles) | off |
| `NEXUS_OPENAI_API_KEY` | Enables `nexus ask`; empty = AI disabled | empty |
| `NEXUS_AI_PROVIDER` | Backend; only `openai` supported | `openai` |
| `NEXUS_AI_MODEL` | Model name for `nexus ask` | `gpt-4o-mini` |
| `NEXUS_AI_BASE_URL` | API endpoint override (gateways/proxies) | `https://api.openai.com/v1` |
| `NEXUS_AI_TIMEOUT_SECS` | Whole-request timeout in seconds | `60` |

No config files, no hardcoded secrets. API keys are never printed.

## AI assistant (optional)

NEXUS works completely without AI. `nexus ask` only activates when
`NEXUS_OPENAI_API_KEY` is set; otherwise it prints setup instructions
and exits 0.

```bash
export NEXUS_OPENAI_API_KEY=<your-key>
./nexus ask "Why might my Docker backend be failing?"
./nexus ask --show-context "Analyze my current system."
```

How it works: CLI → context collector (OS, CPU/memory, Docker, Git,
network, doctor summary — machine facts only, never secrets or file
contents) → provider → LLM → printed answer. The AI is advisory only:
it cannot execute commands and NEXUS never acts on its suggestions
automatically.

## Development commands

```bash
make build   # go build -o nexus .
make run     # go run . --help
make test    # go test ./...
make fmt     # gofmt -w + go vet
make clean   # rm -f nexus
```

## Current feature status (Phase 10 — audited)

| Command | Status |
|---|---|
| `nexus dashboard` | Done — interactive Bubble Tea dashboard (all 7 screens, ↑/↓/Enter/r/Esc/q) |
| `nexus system` | Done — OS/arch/hostname, CPU, memory, disk, uptime via gopsutil |
| `nexus processes` | Done — top-N table (PID/CPU/MEM/MEMORY/NAME/STATUS), `--limit`, `--sort cpu\|memory`, read-only |
| `nexus docker` | Done — read-only `docker ps` table (CONTAINER/IMAGE/STATUS/PORTS/CREATED), `--all`, graceful unavailable message |
| `nexus git` | Done — read-only overview (branch, clean/dirty, modified/staged/untracked, latest commit+date), `--short`, graceful outside repos |
| `nexus network` | Done — read-only diagnostics (internet/DNS/latency + interfaces via stdlib, offline-safe) |
| `nexus doctor` | Done — read-only checks (CPU/memory/disk, network, DNS, Docker, Git, dev ports) with PASS/WARN/FAIL + summary, `--ports` |
| `nexus ask` | Done — optional OpenAI-compatible assistant with env context, `--show-context`, graceful no-key path |

All commands implemented: interactive dashboard plus read-only system, processes,
Docker, Git, network, doctor, and optional AI troubleshooting. See `docs/` for
per-phase reports.
