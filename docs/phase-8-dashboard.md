# Phase 8 Report — Interactive Dashboard

Date: 2026-09-20
Goal: `nexus` (and `nexus dashboard`) launches a Bubble Tea/Lip Gloss dashboard
with Dashboard/System/Processes/Docker/Git/Network/Doctor screens, async loads,
and ↑/↓/Enter/r/Esc/q navigation. AI explicitly out of scope.

## Files created/changed
- `ui/model.go` (new) — `Screen` enum + titles, `Model` (cursor/active/size/
  bodies/loading), `Update` (resize, dataMsg, spec keys), `View` (side-by-side
  menu + bordered content, stacks vertically under 75 cols)
- `ui/load.go` (new) — `LoadScreen()` per screen reusing internal packages
  (zero duplicated collection); dashboard overview composes system/git/docker/
  network/doctor; UI network/doctor probes use 3s timeouts for snappiness
- `ui/styles.go` (new) — Lip Gloss theme, plain variant for NO_COLOR
- `ui/run.go` (new) — `Run(noColor)`; non-TTY returns a clear error, no crash
- `ui/model_test.go` (new) — nav clamping, enter open + load-once, dataMsg,
  refresh reload, Esc home, quit, view chrome, titles
- `ui/doc.go` — kept (package doc now accurate)
- `cmd/root.go` — bare `nexus` launches the dashboard (was: help text)
- `cmd/commands.go` — `dashboardCmd` launches the dashboard (was: stub)
- `cmd/root_test.go` — dashboard excluded from non-interactive run-through
  (interactive path covered by `ui` tests instead)
- `go.mod`/`go.sum` — `bubbletea v1.3.10`, `lipgloss v1.1.0` (+ transitive)
- `README.md` — dashboard usage, keys, status row

## Dependencies
- Added: `bubbletea v1.3.10`, `lipgloss v1.1.0` (plus charm/muesli transitive)

## Commands to test
```bash
go test ./...
./nexus                # interactive dashboard (needs a TTY)
/nexus dashboard       # same
go vet ./...
```

## Tests
- `ui`: 9 tests (navigation, open/load-once, refresh, home, quit, chrome)
- Full suite: `go test ./...` ok (all 9 test packages)
- Manual PTY: renders NEXUS header, menu, live overview (CPU/mem/disk, dirty
  git, 12 containers, network ✓✓, health 7/6/0), `q` exits 0; piped stdin
  fails gracefully (`dashboard needs an interactive terminal`, exit 1)

## Known issues
- No mouse support, no scrolling viewport — long tables (doctor/processes)
  rely on terminal scrollback
- Overview collects sequentially (~2s to full content); shows `Loading…`
  meanwhile, UI never blocks
- `NO_COLOR` disables colors but box-drawing borders remain
