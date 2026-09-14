package system

import (
	"fmt"
	"strings"
	"time"
)

// Format renders Info as aligned human-readable text.
// Unavailable metrics show as "N/A" (never fake values).
func Format(info Info) string {
	var b strings.Builder
	b.WriteString("System\n")
	fmt.Fprintf(&b, "  %-10s %s\n", "OS:", info.OS)
	fmt.Fprintf(&b, "  %-10s %s\n", "Arch:", info.Arch)
	fmt.Fprintf(&b, "  %-10s %s\n", "Hostname:", orNA(info.Hostname, info.HasHost))
	b.WriteString("CPU\n")
	fmt.Fprintf(&b, "  %-10s %s\n", "Model:", orNA(info.CPUModel, info.HasCPU && info.CPUModel != ""))
	fmt.Fprintf(&b, "  %-10s %s\n", "Cores:", intOrNA(info.CPUCores, info.HasCPU))
	fmt.Fprintf(&b, "  %-10s %s\n", "Usage:", percentOrNA(info.CPUPercent, info.HasCPU))
	b.WriteString("Memory\n")
	fmt.Fprintf(&b, "  %-10s %s\n", "Total:", bytesOrNA(info.MemTotal, info.HasMem))
	fmt.Fprintf(&b, "  %-10s %s\n", "Used:", bytesOrNA(info.MemUsed, info.HasMem))
	fmt.Fprintf(&b, "  %-10s %s\n", "Usage:", percentOrNA(info.MemPercent, info.HasMem))
	b.WriteString("Disk (/)\n")
	fmt.Fprintf(&b, "  %-10s %s\n", "Total:", bytesOrNA(info.DiskTotal, info.HasDisk))
	fmt.Fprintf(&b, "  %-10s %s\n", "Used:", bytesOrNA(info.DiskUsed, info.HasDisk))
	fmt.Fprintf(&b, "  %-10s %s\n", "Usage:", percentOrNA(info.DiskPercent, info.HasDisk))
	b.WriteString("Uptime\n")
	fmt.Fprintf(&b, "  %-10s %s\n", "Uptime:", uptimeOrNA(info))
	return b.String()
}

func orNA(s string, ok bool) string {
	if !ok || s == "" {
		return "N/A"
	}
	return s
}

func intOrNA(n int, ok bool) string {
	if !ok || n <= 0 {
		return "N/A"
	}
	return fmt.Sprintf("%d", n)
}

func percentOrNA(p float64, ok bool) string {
	if !ok {
		return "N/A"
	}
	if p < 0 {
		p = 0
	}
	if p > 100 {
		p = 100
	}
	return fmt.Sprintf("%.1f%%", p)
}

func bytesOrNA(n uint64, ok bool) string {
	if !ok {
		return "N/A"
	}
	return formatBytes(n)
}

// formatBytes renders bytes in B/KB/MB/GB/TB with one decimal.
func formatBytes(n uint64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	}
	f := float64(n)
	i := 0
	for f >= 1024 && i < len(units)-1 {
		f /= 1024
		i++
	}
	return fmt.Sprintf("%.1f %s", f, units[i])
}

// formatUptime renders durations like "3d 4h 12m" or "45s".
func formatUptime(d time.Duration) string {
	secs := int64(d.Seconds())
	if secs < 0 {
		secs = 0
	}
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	days := secs / 86400
	hours := (secs % 86400) / 3600
	mins := (secs % 3600) / 60
	var parts []string
	if days > 0 {
		parts = append(parts, fmt.Sprintf("%dd", days))
	}
	if hours > 0 || days > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
	}
	parts = append(parts, fmt.Sprintf("%dm", mins))
	return strings.Join(parts, " ")
}

func uptimeOrNA(info Info) string {
	if !info.HasUptime {
		return "N/A"
	}
	return formatUptime(info.Uptime)
}
