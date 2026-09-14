# Phase 2 Report — System Information

Date: 2026-09-20
Goal: Real `nexus system` output (OS, CPU, memory, disk, uptime) via gopsutil.
No fakes, no unrelated command changes, no dashboard/processes.

## Files created/changed
- `internal/system/system.go` (new) — `Info` struct with `Has*` availability flags,
  testable `provider` + `collectWith`, `Collect()` joining per-metric errors via `errors.Join`
- `internal/system/format.go` (new) — `Format()`, `formatBytes()`, `formatUptime()`,
  `percentOrNA()`; unavailable metrics render `N/A`
- `internal/system/system_test.go` (new) — stub-provider collect tests, format tests,
  byte/uptime/percent unit tests
- `cmd/commands.go` — `systemCmd` now calls `system.Collect()` + prints `system.Format()`;
  partial errors go to stderr as `Warning:`, stdout stays clean; all other commands untouched
- `go.mod`/`go.sum` — added `github.com/shirou/gopsutil/v4 v4.26.8`
- `README.md` — `nexus system` row marked done

## Dependencies
- Added: `github.com/shirou/gopsutil/v4 v4.26.8` (+ transitive `golang.org/x/sys`, `lufia/plan9stats`, `tklauser/go-sysconf`, `power-devops/perfstat`, `ebitengine/purego`, `tklauser/numcpus`)
- Existing: `github.com/spf13/cobra v1.10.2`

## Commands to test
```bash
go test ./...
go run . system
go vet ./...
```

## Tests
- `internal/system`: `TestCollectWithStubSuccess` (stub provider → all fields set, no error),
  `TestCollectWithPartialFailure` (hostname+mem fail → flags false, joined error, rest intact),
  `TestFormatShowsAllFields` (all 13 required values present), `TestFormatNAOnEmpty`,
  `TestFormatBytes` (B→TB), `TestFormatUptime` (45s→3d 4h 12m), `TestPercentOrNA` (clamping)
- Full suite: `go test ./...` ok (cmd, config, logger, system)
- Manual `go run . system` on macOS arm64: darwin/arm64, Apple M3, 8 cores, real %/GB/uptime

## Known issues
- CPU usage uses a 500ms sample (`cpu.Percent`) so `nexus system` takes ~0.5s; acceptable for MVP
- Disk reports `/` only; multi-volume breakdown deferred
- `cmd/root_test.go` now exercises the real collector (~0.5s per system invocation)
