package git

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strings"
	"time"
)

// Repo is a read-only snapshot of a Git working tree.
type Repo struct {
	Root     string
	Branch   string
	Detached bool

	Clean     bool
	Modified  []string
	Staged    []string
	Untracked []string

	CommitHash    string
	CommitSubject string
	CommitDate    string
}

// NotRepoError means dir is not inside a Git working tree.
type NotRepoError struct{ Dir string }

func (e *NotRepoError) Error() string { return "not a git repository: " + e.Dir }

// commandTimeout bounds git calls so a hung repo cannot hang NEXUS.
const commandTimeout = 10 * time.Second

// gitCLI is the seam to the git binary (real runner vs test stub).
type gitCLI interface {
	run(dir string, args ...string) (stdout, stderr string, err error)
}

type runner struct{}

func (runner) run(dir string, args ...string) (string, string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	if dir != "" {
		cmd.Dir = dir
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

// Inspect snapshots the repository containing dir ("" = current directory).
// It runs only read-only commands and never modifies the repository.
func Inspect(dir string) (Repo, error) {
	return inspectWith(runner{}, dir)
}

func inspectWith(cli gitCLI, dir string) (Repo, error) {
	var repo Repo

	root, _, err := cli.run(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		if dir == "" {
			dir = "current directory"
		}
		return Repo{}, &NotRepoError{Dir: dir}
	}
	repo.Root = strings.TrimSpace(root)

	branch, _, err := cli.run(dir, "branch", "--show-current")
	if err == nil {
		repo.Branch = strings.TrimSpace(branch)
	}
	if repo.Branch == "" {
		// Detached HEAD (or unborn branch): fall back to short SHA.
		if sha, _, shaErr := cli.run(dir, "rev-parse", "--short", "HEAD"); shaErr == nil {
			repo.Branch = strings.TrimSpace(sha)
			repo.Detached = true
		} else {
			repo.Branch = "(no commits yet)"
		}
	}

	statusOut, _, err := cli.run(dir, "status", "--porcelain=v1", "-z")
	if err != nil {
		return repo, errors.New("git status failed: " + shortErr(err))
	}
	repo.Modified, repo.Staged, repo.Untracked = parsePorcelainZ(statusOut)
	repo.Clean = len(repo.Modified) == 0 && len(repo.Staged) == 0 && len(repo.Untracked) == 0

	if logOut, _, logErr := cli.run(dir, "log", "-1", "--format=%H%x00%h%x00%s%x00%cI"); logErr == nil {
		if parts := strings.Split(strings.TrimSuffix(logOut, "\n"), "\x00"); len(parts) == 4 {
			repo.CommitHash, repo.CommitSubject, repo.CommitDate =
				strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2]), strings.TrimSpace(parts[3])
		}
	}
	// A repo with no commits has no log; that is valid, not an error.

	return repo, nil
}

func shortErr(err error) string {
	if s := strings.TrimSpace(err.Error()); s != "" {
		return s
	}
	return "unknown error"
}

// parsePorcelainZ parses `git status --porcelain=v1 -z` output.
// Entries are NUL-separated "XY path" (renames: "XY orig\0new").
// Staged = index status (X) is not ' ' or '?'; modified = worktree (Y)
// is not ' ' or '?'; untracked ("??") paths go to untracked.
func parsePorcelainZ(output string) (modified, staged, untracked []string) {
	records := strings.Split(output, "\x00")
	for i := 0; i < len(records); i++ {
		rec := records[i]
		if rec == "" {
			continue
		}
		if len(rec) < 4 || rec[2] != ' ' {
			continue // malformed; skip instead of crashing
		}
		x, y, path := rec[0], rec[1], rec[3:]
		if x == '?' && y == '?' {
			untracked = append(untracked, path)
			continue
		}
		if x == 'R' || y == 'R' {
			// Rename entries carry "orig\0new": attribute to the new path.
			if i+1 < len(records) && records[i+1] != "" {
				i++
				path = records[i]
			}
		}
		if x != ' ' {
			staged = append(staged, path)
		}
		if y != ' ' {
			modified = append(modified, path)
		}
	}
	return modified, staged, untracked
}

// NotRepoMessage is the user-facing text outside a Git working tree.
func NotRepoMessage() string {
	return "Not a git repository.\nRun this command inside a Git working tree.\n"
}
