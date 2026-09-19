package cmd

import (
	"bytes"
	"testing"
)

func TestAllCommandsRegistered(t *testing.T) {
	want := []string{"dashboard", "system", "processes", "docker", "git", "network", "doctor", "ask"}
	got := map[string]bool{}
	for _, c := range rootCmd.Commands() {
		got[c.Name()] = true
	}
	for _, w := range want {
		if !got[w] {
			t.Errorf("missing command %q", w)
		}
	}
}

func TestNonInteractiveCommandsDoNotError(t *testing.T) {
	// dashboard is interactive (needs a TTY) and is covered by ui tests.
	for _, name := range []string{"system", "processes", "docker", "git", "network", "doctor"} {
		rootCmd.SetOut(new(bytes.Buffer))
		rootCmd.SetArgs([]string{name})
		if err := rootCmd.Execute(); err != nil {
			t.Errorf("%s: unexpected error: %v", name, err)
		}
	}
	// ask requires an argument
	rootCmd.SetOut(new(bytes.Buffer))
	rootCmd.SetArgs([]string{"ask", "test question"})
	if err := rootCmd.Execute(); err != nil {
		t.Errorf("ask: unexpected error: %v", err)
	}
}
