package docker

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// FormatTable renders containers as an aligned terminal table.
// Zero containers renders the header plus a "(no containers)" note.
func FormatTable(containers []Container) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "CONTAINER\tIMAGE\tSTATUS\tPORTS\tCREATED")
	if len(containers) == 0 {
		fmt.Fprintln(w, "(no containers)")
	} else {
		for _, c := range containers {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				truncate(c.Name, 32),
				truncate(c.Image, 40),
				c.Status,
				orDash(c.Ports),
				c.Created,
			)
		}
	}
	_ = w.Flush()
	return b.String()
}

func orDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 1 {
		return s[:max]
	}
	return s[:max-1] + "…"
}
