# Phase 5 Report — Git Inspector

Date: 2026-09-20
Goal: Read-only `nexus git` overview (repository, branch, clean/dirty, modified,
staged, untracked, latest commit + date) via the Git CLI. Graceful outside repos,
`--short` one-liner, never modifies. No new dependencies.

## Files created/changed
- `internal/git/git.go` (new) — `Repo`, `Inspect(dir)` running only
  `rev-parse`, `branch --show-current`, `status --porcelain=v1 -z`, `log -1`
  (10s timeout); `NotRepoError` for non-repos; `parsePorcelainZ()` with rename
  (`R orig\0new`) and malformed-line handling; `gitCLI` interface seam
- `internal/git/format.go` (new) — `Format()` (labeled multi-line view),
  `FormatShort()` (`<branch>[*] <hash> <subject> (M/S/U counts)`), 10-file list cap
- `internal/git/git_test.go` (new) — porcelain parse (staged/modified/untracked/
  rename), empty + garbage input, stubbed inspect success / not-a-repo / detached
  HEAD, full + clean format, short format incl. one-line assertion, message text
- `cmd/commands.go` — `gitCmd` wired: `Inspect("")`, non-repo → message exit 0,
  `--short` flag; other commands untouched
- `README.md` — `nexus git` row marked done

## Dependencies
- None new (stdlib `os/exec` + `context`; git CLI is an external binary)

## Commands to test
```bash
go test ./...
go run . git
go run . git --short
cd /tmp && go run /Users/viru/Desktop/projects/nexus git   # non-repo path
go vet ./...
```

## Tests
- `internal/git`: porcelain parsing incl. renames, stubbed success/detached/
  not-a-repo branches, format content + clean + short one-liner
- Full suite: `go test ./...` ok (cmd, config, docker, git, logger, process, system)
- Manual: live repo shows main/dirty with correct file lists + commit `8050562`;
  `--short` one line; `/tmp` run prints `Not a git repository.` with exit 0

## Known issues
- Works on the current directory only; no `--path` flag (deferred)
- File lists capped at 10 entries in full view (`+N more`)
- Detached HEAD shows short SHA with `(detached)` marker; unborn branches show `(no commits yet)`
