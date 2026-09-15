package process

import (
	"fmt"
	"strings"
	"text/tabwriter"
)

// FormatTable renders procs as an aligned terminal table.
// Empty input renders the header plus a "(no processes found)" note.
func FormatTable(procs []Proc) string {
	var b strings.Builder
	w := tabwriter.NewWriter(&b, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "PID\tCPU%\tMEM%\tMEMORY\tNAME\tSTATUS")
	if len(procs) == 0 {
		fmt.Fprintln(w, "(no processes found)")
	} else {
		for _, p := range procs {
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\t%s\t%s\n",
				p.PID,
				formatPercent(p.CPUPercent),
				formatPercent(p.MemPercent),
				formatBytes(p.RSS),
				truncate(p.Name, 32),
				p.Status,
			)
		}
	}
	_ = w.Flush()
	return b.String()
}

func formatPercent(p float64) string {
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return fmt.Sprintf("%.1f%%", p)
}

// formatBytes renders RSS in B/KB/MB/GB with one decimal.
func formatBytes(n uint64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	f, units := float64(n), []string{"B", "KB", "MB", "GB", "TB"}
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", f, units[i])
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
