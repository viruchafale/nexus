package docker

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const samplePs = "web-1\tnginx:1.25\tUp 3 hours\t0.0.0.0:8080->80/tcp\t2026-09-19 10:00:00 +0000 UTC\n" +
	"db-1\tpostgres:16\tExited (0) 2 days ago\t\t2026-09-10 08:00:00 +0000 UTC\n"

func TestParsePs(t *testing.T) {
	got := parsePs(samplePs)
	if len(got) != 2 {
		t.Fatalf("got %d containers, want 2: %v", len(got), got)
	}
	if got[0].Name != "web-1" || got[0].Image != "nginx:1.25" || got[0].Status != "Up 3 hours" {
		t.Errorf("first = %+v", got[0])
	}
	if got[0].Ports != "0.0.0.0:8080->80/tcp" {
		t.Errorf("ports = %q", got[0].Ports)
	}
	if got[1].Ports != "" {
		t.Errorf("empty ports should stay empty pre-format, got %q", got[1].Ports)
	}
}

func TestParsePsSkipsGarbage(t *testing.T) {
	got := parsePs("\n\nnot-a-valid-line\n" + samplePs + "too\tfew\n")
	if len(got) != 2 {
		t.Errorf("should skip blank/malformed lines, got %v", got)
	}
	if got := parsePs(""); len(got) != 0 {
		t.Errorf("empty output should parse to zero containers, got %v", got)
	}
}

// stubRunner fakes the docker CLI without a daemon.
type stubRunner struct {
	pathErr error
	stdout  string
	stderr  string
	runErr  error
	sawAll  bool
}

func (s *stubRunner) lookPath() (string, error) {
	if s.pathErr != nil {
		return "", s.pathErr
	}
	return "/usr/bin/docker", nil
}

func (s *stubRunner) ps(_ context.Context, all bool) (string, string, error) {
	s.sawAll = all
	return s.stdout, s.stderr, s.runErr
}

func TestListNotInstalled(t *testing.T) {
	_, err := listWith(&stubRunner{pathErr: errors.New("executable not found")}, false)
	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Fatalf("expected UnavailableError, got %v", err)
	}
}

func TestListDaemonDown(t *testing.T) {
	_, err := listWith(&stubRunner{
		runErr: errors.New("exit status 1"),
		stderr: "Cannot connect to the Docker daemon",
	}, false)
	var unavail *UnavailableError
	if !errors.As(err, &unavail) {
		t.Fatalf("expected UnavailableError, got %v", err)
	}
	if !strings.Contains(unavail.Reason, "Cannot connect") {
		t.Errorf("reason should carry daemon stderr, got %q", unavail.Reason)
	}
}

func TestListPassesAllFlag(t *testing.T) {
	r := &stubRunner{stdout: samplePs}
	got, err := listWith(r, true)
	if err != nil {
		t.Fatal(err)
	}
	if !r.sawAll {
		t.Error("--all should reach `docker ps -a`")
	}
	if len(got) != 2 {
		t.Errorf("got %d containers", len(got))
	}
}

func TestUnavailableMessage(t *testing.T) {
	msg := UnavailableMessage()
	for _, want := range []string{
		"Docker is not available.",
		"- Docker is not installed",
		"- Docker daemon is not running",
		"- Docker socket is unavailable",
	} {
		if !strings.Contains(msg, want) {
			t.Errorf("message missing %q\n%s", want, msg)
		}
	}
}

func TestFormatTable(t *testing.T) {
	out := FormatTable(parsePs(samplePs))
	for _, want := range []string{"CONTAINER", "IMAGE", "STATUS", "PORTS", "CREATED", "web-1", "nginx:1.25", "Up 3 hours", "8080->80", "db-1", "-"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\n%s", want, out)
		}
	}
}

func TestFormatTableEmpty(t *testing.T) {
	out := FormatTable(nil)
	if !strings.Contains(out, "CONTAINER") || !strings.Contains(out, "no containers") {
		t.Errorf("empty table should show header + note\n%s", out)
	}
}
