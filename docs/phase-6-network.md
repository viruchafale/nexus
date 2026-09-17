# Phase 6 Report — Network Diagnostics

Date: 2026-09-20
Goal: Read-only `nexus network` report (internet, DNS, latency, interfaces,
hostname) using stdlib only. Offline-safe, no scans, no privileged ops,
configurable timeouts. No new dependencies.

## Files created/changed
- `internal/network/network.go` (new) — `Config`/`DefaultConfig` (5s DNS + dial
  timeouts, `example.com` DNS target, `1.1.1.1:443` dial target), `Collect()`/
  `CollectWith()`, `listInterfaces()` (skips loopback, keeps down ifaces),
  `addrIP()` (skips loopback + link-local IPv6); at most one DNS lookup + one
  TCP dial, no listeners, no scans, nothing modified
- `internal/network/format.go` (new) — `Format()` with NETWORK (✓/✗, latency)
  and INTERFACES sections + hostname; `formatDuration()`, down/empty handling
- `internal/network/network_test.go` (new) — online/offline report content,
  empty interfaces, duration cases, IP-list cases, loopback-exclusion on the
  real machine (network-proof, no external dependency)
- `cmd/commands.go` — `networkCmd` wired to `Collect()` + `Format()`; never
  errors (offline = ✗/N/A report); other commands untouched
- `README.md` — `nexus network` row marked done

## Dependencies
- None new (stdlib `net`, `os`, `context`, `time` only)

## Commands to test
```bash
go test ./...
go run . network
go vet ./...
```

## Tests
- `internal/network`: online report (✓, `24 ms`, iface IPs, `(down)`, hostname),
  offline report (✗, N/A, detail line, full sections intact), no-iface note,
  duration/edge cases, loopback exclusion on localhost
- Full suite: `go test ./...` ok (all packages)
- Manual: live run shows Internet ✓, DNS ✓, 349 ms, en0 real IP, hostname

## Known issues
- Internet check dials `1.1.1.1:443`; networks blocking it report offline
  (DNS section still independently valid)
- Latency is single-sample TCP connect time, not a sustained measurement
- Up interfaces with only link-local addresses show `-` (by design, avoids noise)
