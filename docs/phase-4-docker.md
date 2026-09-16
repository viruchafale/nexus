# Phase 4 Report — Docker Inspection

Date: 2026-09-20
Goal: Read-only `nexus docker` table (CONTAINER, IMAGE, STATUS, PORTS, CREATED)
via the Docker CLI. Graceful unavailable message, `--all` flag, no container
management. No new dependencies.

## Files created/changed
- `internal/docker/docker.go` (new) — `Container`, `List(all)` (runs only
  `docker ps --format`, 10s timeout via context), `UnavailableError`
  (not-installed vs daemon-unreachable with daemon stderr as reason),
  `parsePs()`, `UnavailableMessage()`, `Detail()`; `dockerCLI` interface seam
  (real `runner` vs test stub)
- `internal/docker/format.go` (new) — `FormatTable()` via `text/tabwriter`,
  empty ports render `-`, long names/images truncated
- `internal/docker/docker_test.go` (new) — parse (fields, empty ports),
  garbage-line skipping, empty output, not-installed and daemon-down via stub,
  `--all` passthrough, unavailable-message wording, table content, empty table
- `cmd/commands.go` — `dockerCmd` wired: `List()`, unavailable → spec message
  on stdout + detail on stderr, exit 0; `--all`/`-a` flag; other commands untouched
- `README.md` — `nexus docker` row marked done

## Dependencies
- None new (stdlib `os/exec` + `context` only; Docker CLI is an external binary)

## Commands to test
```bash
go test ./...
go run . docker
go run . docker --all
PATH=/usr/bin:/bin go run . docker   # simulate missing Docker → unavailable message, exit 0
go vet ./...
```

## Tests
- `internal/docker`: parse real-format sample, skip blank/malformed lines,
  `UnavailableError` for missing binary and daemon-down (reason carries daemon
  stderr), `--all` reaches `ps`, message matches spec wording, table shows all
  columns + `-` for empty ports, empty list shows `(no containers)`
- Full suite: `go test ./...` ok (cmd, config, docker, logger, process, system)
- Manual: live daemon lists 12 containers; missing-binary run prints exactly the
  spec unavailable message with exit 0

## Known issues
- Requires the `docker` CLI on PATH; Docker API (socket) support deferred
- Only `docker ps` columns shown; image lists, stats, and logs are out of scope
- 10s CLI timeout guards hung daemons but adds worst-case latency
