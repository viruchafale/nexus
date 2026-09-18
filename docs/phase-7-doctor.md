# Phase 7 Report — Nexus Doctor

Date: 2026-09-20
Goal: Read-only `nexus doctor` engine — independent PASS/WARN/FAIL checks with
summary, warnings, and recommendations. Zero auto-fix. No new dependencies.

## Files created/changed
- `internal/doctor/doctor.go` (new) — `Status`/`Result`/`Check` abstraction,
  `Run()` (CPU, memory, disk, network, DNS, Docker, Git + 6 dev ports);
  reuses `internal/system`, `network`, `git`, `docker` (no logic duplicated);
  thresholds (CPU 70/90, mem 80/95, disk 80/90); 3s doctor network timeouts;
  port occupancy via localhost TCP dial only
- `internal/doctor/format.go` (new) — spec layout (`NEXUS DOCTOR`, ✓/⚠/✗ rows,
  SUMMARY tally, Warnings + Recommendations sections omitted when all pass)
- `internal/doctor/doctor_test.go` (new) — threshold edges + unavailable→WARN,
  real loopback listener for occupied port, tally, full/spec-assertion format,
  all-pass section omission, check shape (name + detail on non-Pass)
- `cmd/commands.go` — `doctorCmd` wired to `Run()` + `Format()`; never errors;
  other commands untouched
- `README.md` — `nexus doctor` row marked done

## Dependencies
- None new (stdlib `net`/`time` + existing internal packages)

## Commands to test
```bash
go test ./...
go run . doctor
go run . doctor --ports 5432,6379
go run . doctor --ports abc   # expect Error, exit 1
go vet ./...
```

## Tests
- `internal/doctor`: 7 threshold cases, occupied-port via real listener,
  tally 2/1/1, format contains every spec element, all-pass omits sections
- Full suite: `go test ./...` ok (all 8 test packages)
- Manual: live run matches spec layout — 7 passed, 6 warnings (mem/disk ~84%,
  dirty git, ports 3000/5173/6379 occupied), 0 failed, with recommendations

## Known issues (all fixed post-phase)
- ~~`system.Collect()` runs 3×~~ Fixed: `RunWith(Options)` collects one system
  snapshot shared by CPU/memory/disk result constructors (pure + unit-tested);
  full `doctor` run is now ~1.9s. Checks stay independent as separate Results.
- ~~Doctor uses shorter 3s timeouts~~ Fixed: doctor uses `network.DefaultConfig()`
  (5s), identical to `nexus network`; `Options.Net` still allows overrides.
- ~~Fixed port list~~ Fixed: `nexus doctor --ports 3000,5432` via `ParsePorts()`
  (1-65535 validated, explicit errors); empty means DevPorts.
