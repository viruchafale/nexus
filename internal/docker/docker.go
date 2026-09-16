package docker

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

// Container is a read-only snapshot of one Docker container.
type Container struct {
	Name    string
	Image   string
	Status  string
	Ports   string
	Created string
}

// UnavailableError explains why Docker cannot be inspected.
// It is returned (with no containers) when Docker is missing
// or the daemon is unreachable — never a crash.
type UnavailableError struct {
	Reason string // "not installed" or daemon stderr
}

func (e *UnavailableError) Error() string { return "docker unavailable: " + e.Reason }

// psFormat uses tab separators: names, image, status, ports, created.
const psFormat = "{{.Names}}\t{{.Image}}\t{{.Status}}\t{{.Ports}}\t{{.CreatedAt}}"

// commandTimeout bounds the docker CLI call so a hung daemon
// cannot hang NEXUS (fast startup principle).
const commandTimeout = 10 * time.Second

// List returns running containers (all=true includes stopped ones).
// It runs only `docker ps` — never start/stop/delete/restart.
func List(all bool) ([]Container, error) {
	return listWith(runner{}, all)
}

// dockerCLI is the seam to the docker binary; runner is the real
// implementation, tests substitute a stub. Kept as a small interface
// so List stays testable without a daemon or extra dependencies.
type dockerCLI interface {
	lookPath() (string, error)
	ps(ctx context.Context, all bool) (stdout, stderr string, err error)
}

// runner executes the real docker CLI.
type runner struct{}

func (runner) lookPath() (string, error) { return exec.LookPath("docker") }

func (runner) ps(ctx context.Context, all bool) (string, string, error) {
	args := []string{"ps", "--no-trunc", "--format", psFormat}
	if all {
		args = append(args, "-a")
	}
	cmd := exec.CommandContext(ctx, "docker", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	return stdout.String(), stderr.String(), err
}

func listWith(cli dockerCLI, all bool) ([]Container, error) {
	if _, err := cli.lookPath(); err != nil {
		return nil, &UnavailableError{Reason: "docker CLI not installed"}
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()
	stdout, stderr, err := cli.ps(ctx, all)
	if err != nil {
		reason := strings.TrimSpace(stderr)
		if reason == "" {
			reason = err.Error()
		}
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			reason = "docker CLI timed out after " + commandTimeout.String()
		}
		return nil, &UnavailableError{Reason: reason}
	}
	return parsePs(stdout), nil
}

// parsePs parses tab-separated `docker ps --format` output.
// Malformed lines are skipped; an empty daemon response is valid (no containers).
func parsePs(output string) []Container {
	var out []Container
	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 5 {
			continue
		}
		out = append(out, Container{
			Name:    fields[0],
			Image:   fields[1],
			Status:  fields[2],
			Ports:   fields[3],
			Created: fields[4],
		})
	}
	return out
}

// UnavailableMessage is the user-facing text when Docker cannot be reached.
func UnavailableMessage() string {
	return `Docker is not available.

Possible reasons:
- Docker is not installed
- Docker daemon is not running
- Docker socket is unavailable
`
}

// Detail formats the underlying cause for stderr (keeps stdout scriptable).
func Detail(err error) string {
	var unavail *UnavailableError
	if errors.As(err, &unavail) {
		return fmt.Sprintf("Detail: %s", unavail.Reason)
	}
	return fmt.Sprintf("Detail: %s", err)
}
