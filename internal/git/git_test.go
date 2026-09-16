package git

import (
	"errors"
	"strings"
	"testing"
)

func TestParsePorcelainZ(t *testing.T) {
	// M staged-new, M modified, ?? untracked, A staged, R rename staged.
	out := "M  new.txt\x00 M old.txt\x00?? untracked.txt\x00A  added.txt\x00R  orig.txt\x00new-name.txt\x00"
	mod, staged, untracked := parsePorcelainZ(out)

	if len(mod) != 1 || mod[0] != "old.txt" {
		t.Errorf("modified = %v", mod)
	}
	wantStaged := map[string]bool{"new.txt": true, "added.txt": true, "new-name.txt": true}
	if len(staged) != len(wantStaged) {
		t.Errorf("staged = %v", staged)
	}
	for _, s := range staged {
		if !wantStaged[s] {
			t.Errorf("unexpected staged %q in %v", s, staged)
		}
	}
	if len(untracked) != 1 || untracked[0] != "untracked.txt" {
		t.Errorf("untracked = %v", untracked)
	}
}

func TestParsePorcelainZEmptyAndGarbage(t *testing.T) {
	if mod, st, un := parsePorcelainZ(""); len(mod)+len(st)+len(un) != 0 {
		t.Errorf("empty should parse to nothing, got %v %v %v", mod, st, un)
	}
	// Malformed records are skipped, never crash.
	if mod, st, un := parsePorcelainZ("bogus\x00X\x00"); len(mod)+len(st)+len(un) != 0 {
		t.Errorf("garbage should parse to nothing, got %v %v %v", mod, st, un)
	}
}

// stubCLI fakes the git binary.
type stubCLI struct {
	outputs map[string]string
	fail    map[string]bool
}

func (s *stubCLI) run(_ string, args ...string) (string, string, error) {
	key := strings.Join(args, " ")
	if s.fail[key] {
		return "", "fatal: not a git repository", errors.New("exit status 128")
	}
	return s.outputs[key], "", nil
}

func okStub() *stubCLI {
	return &stubCLI{outputs: map[string]string{
		"rev-parse --show-toplevel":             "/repo\n",
		"branch --show-current":                 "main\n",
		"status --porcelain=v1 -z":              " M a.txt\x00?? b.txt\x00",
		"log -1 --format=%H%x00%h%x00%s%x00%cI": "abc123full\x00abc1234\x00Add thing\x002026-09-20T10:00:00+05:30",
	}}
}

func TestInspectSuccess(t *testing.T) {
	repo, err := inspectWith(okStub(), "")
	if err != nil {
		t.Fatal(err)
	}
	if repo.Root != "/repo" || repo.Branch != "main" {
		t.Errorf("got %+v", repo)
	}
	if repo.Clean {
		t.Error("should be dirty with modified + untracked files")
	}
	if len(repo.Modified) != 1 || len(repo.Untracked) != 1 || len(repo.Staged) != 0 {
		t.Errorf("got %+v", repo)
	}
	if repo.CommitHash != "abc1234" || repo.CommitSubject != "Add thing" {
		t.Errorf("commit = %+v", repo)
	}
}

func TestInspectNotARepo(t *testing.T) {
	stub := okStub()
	stub.fail = map[string]bool{"rev-parse --show-toplevel": true}
	_, err := inspectWith(stub, "/tmp")
	var notRepo *NotRepoError
	if !errors.As(err, &notRepo) {
		t.Fatalf("expected NotRepoError, got %v", err)
	}
}

func TestInspectDetached(t *testing.T) {
	stub := okStub()
	stub.outputs["branch --show-current"] = "\n"
	stub.outputs["rev-parse --short HEAD"] = "deadbee\n"
	repo, err := inspectWith(stub, "")
	if err != nil {
		t.Fatal(err)
	}
	if !repo.Detached || repo.Branch != "deadbee" {
		t.Errorf("got %+v", repo)
	}
}

func TestFormat(t *testing.T) {
	repo, _ := inspectWith(okStub(), "")
	repo.Root = "/repo"
	out := Format(repo)
	for _, want := range []string{"/repo", "main", "dirty", "a.txt", "b.txt", "abc1234", "Add thing", "2026-09-20"} {
		if !strings.Contains(out, want) {
			t.Errorf("Format missing %q\n%s", want, out)
		}
	}
}

func TestFormatClean(t *testing.T) {
	out := Format(Repo{Root: "/r", Branch: "main", Clean: true})
	if !strings.Contains(out, "clean") || !strings.Contains(out, "none") {
		t.Errorf("clean repo should show clean/none\n%s", out)
	}
}

func TestFormatShort(t *testing.T) {
	repo, _ := inspectWith(okStub(), "")
	out := FormatShort(repo)
	for _, want := range []string{"main*", "abc1234", "Add thing", "M1 S0 U1"} {
		if !strings.Contains(out, want) {
			t.Errorf("short missing %q: %q", want, out)
		}
	}
	if n := strings.Count(strings.TrimSpace(out), "\n"); n != 0 {
		t.Errorf("short should be one line, got %q", out)
	}
}

func TestNotRepoMessage(t *testing.T) {
	if !strings.Contains(NotRepoMessage(), "Not a git repository.") {
		t.Errorf("got %q", NotRepoMessage())
	}
}
