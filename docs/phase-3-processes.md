# Phase 3 Report — Process Monitor

Date: 2026-09-20
Goal: Read-only `nexus processes` table (PID, CPU%, MEM%, MEMORY, NAME, STATUS)
via gopsutil. Sorted by CPU by default, top 15, with `--limit`/`--sort` flags.
No killing, no unrelated command changes.

## Files created/changed
- `internal/process/process.go` (new) — `Proc` snapshot, `Collect()` (prime + 200ms
  sample, per-process errors skipped, fatal only if listing fails), `Sort()`,
  `Top()`, `ParseSortMode()`, `SortError`/`LimitError` (explicit errors)
- `internal/process/format.go` (new) — `FormatTable()` via `text/tabwriter`,
  percent/bytes helpers, 32-char name truncation
- `internal/process/process_test.go` (new) — CPU/mem sort order, sort-mode parsing
  incl. invalid values, Top limits incl. 0/negative, table content, empty table, truncate
- `cmd/commands.go` — `processesCmd` wired: parse `--sort`, `Collect()`,
  `Sort()`, `Top()`, print table; `--limit` (default 15), `--sort` (default cpu);
  all other commands untouched
- `README.md` — `nexus processes` row marked done

## Dependencies
- No new deps (reuses `github.com/shirou/gopsutil/v4 v4.26.8`)

## Commands to test
```bash
go test ./...
go run . processes
go run . processes --limit 20
go run . processes --sort memory --limit 5
go run . processes --sort disk    # expect: Error: invalid --sort value
go run . processes --limit 0      # expect: Error: invalid --limit
go vet ./...
```

## Tests
- `internal/process`: sort CPU (`2,4,3,1`) and mem (`3,4,1,2`) orders on synthetic
  data, `ParseSortMode` accepts cpu/mem/memory (case-insensitive) and rejects
  empty/disk/pid, `Top` slices and rejects <1, table contains header + names + `2.0 GB`,
  empty input shows `(no processes found)`, truncate caps at 32 runes
- Full suite: `go test ./...` ok (cmd, config, logger, process, system)
- Manual on macOS arm64: real table, both sorts, bad flags exit 1 with clear stderr errors

## Known issues
- Per-process CPU% uses a single 200ms prime+sample window, so values are
  approximate (standard MVP tradeoff; sorting still meaningful)
- First-run CPU% for very short-lived processes may read 0.0%
- STATUS strings come from gopsutil verbatim (e.g. `sleep` on macOS)
- Table is a one-shot snapshot, not a live/interactive view (dashboard phase may add that)
