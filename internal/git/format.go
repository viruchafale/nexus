package git

import (
	"fmt"
	"strings"
)

// Format renders a concise multi-line repository overview.
func Format(r Repo) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-14s %s\n", "Repository:", r.Root)
	branch := r.Branch
	if r.Detached {
		branch += " (detached)"
	}
	fmt.Fprintf(&b, "%-14s %s\n", "Branch:", branch)
	status := "clean"
	if !r.Clean {
		status = "dirty"
	}
	fmt.Fprintf(&b, "%-14s %s\n", "Status:", status)
	fmt.Fprintf(&b, "%-14s %s\n", "Modified:", listOrNone(r.Modified))
	fmt.Fprintf(&b, "%-14s %s\n", "Staged:", listOrNone(r.Staged))
	fmt.Fprintf(&b, "%-14s %s\n", "Untracked:", listOrNone(r.Untracked))
	if r.CommitHash == "" {
		fmt.Fprintf(&b, "%-14s %s\n", "Commit:", "(no commits yet)")
	} else {
		fmt.Fprintf(&b, "%-14s %s  %s\n", "Commit:", r.CommitHash, r.CommitSubject)
		fmt.Fprintf(&b, "%-14s %s\n", "Date:", r.CommitDate)
	}
	return b.String()
}

// FormatShort renders a one-line summary for `nexus git --short`.
func FormatShort(r Repo) string {
	mark := ""
	if !r.Clean {
		mark = "*"
	}
	commit := r.CommitHash
	if commit == "" {
		commit = "-"
	}
	subject := r.CommitSubject
	if subject == "" {
		subject = "(no commits yet)"
	}
	return fmt.Sprintf("%s%s %s %s (M%d S%d U%d)\n",
		r.Branch, mark, commit, subject,
		len(r.Modified), len(r.Staged), len(r.Untracked))
}

func listOrNone(files []string) string {
	if len(files) == 0 {
		return "none"
	}
	const max = 10
	shown := files
	more := ""
	if len(files) > max {
		shown = files[:max]
		more = fmt.Sprintf(" (+%d more)", len(files)-max)
	}
	return strings.Join(shown, ", ") + more
}
