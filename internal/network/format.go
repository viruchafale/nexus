package network

import (
	"fmt"
	"strings"
	"time"
)

// Format renders the diagnostics report: a NETWORK status section
// followed by an INTERFACES address section.
func Format(info NetInfo) string {
	var b strings.Builder
	b.WriteString("NETWORK\n")
	b.WriteString("────────────────────────\n")
	fmt.Fprintf(&b, "%-12s %s\n", "Internet", mark(info.NetOK))
	fmt.Fprintf(&b, "%-12s %s\n", "DNS", mark(info.DNSOK))
	fmt.Fprintf(&b, "%-12s %s\n", "Latency", formatLatency(info))
	if info.NetDetail != "" && !info.NetOK {
		fmt.Fprintf(&b, "%-12s %s\n", "Note", info.NetDetail)
	}
	b.WriteString("\nINTERFACES\n")
	b.WriteString("────────────────────────\n")
	if len(info.Interfaces) == 0 {
		b.WriteString("(no interfaces found)\n")
	} else {
		for _, iface := range info.Interfaces {
			fmt.Fprintf(&b, "%-12s %s\n", iface.Name, formatIPs(iface))
		}
	}
	if info.Hostname != "" && info.HasHost {
		fmt.Fprintf(&b, "\nHostname: %s\n", info.Hostname)
	}
	return b.String()
}

func mark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

// formatLatency shows measured dial time, or N/A offline.
func formatLatency(info NetInfo) string {
	if !info.HasLatency {
		return "N/A"
	}
	return formatDuration(info.Latency)
}

func formatDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	ms := float64(d) / float64(time.Millisecond)
	if ms < 1000 {
		return fmt.Sprintf("%.0f ms", ms)
	}
	return fmt.Sprintf("%.1f s", ms/1000)
}

func formatIPs(iface IfaceInfo) string {
	if len(iface.IPs) == 0 {
		if !iface.Up {
			return "(down)"
		}
		return "-"
	}
	return strings.Join(iface.IPs, ", ")
}
