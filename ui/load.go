package ui

import (
	"fmt"
	"strings"
	"time"

	"nexus/internal/docker"
	"nexus/internal/doctor"
	"nexus/internal/git"
	"nexus/internal/network"
	"nexus/internal/process"
	"nexus/internal/system"
)

// quickNetConfig keeps dashboard/network loads snappy: the full
// `nexus network` command uses 5s timeouts, the UI uses 3s.
func quickNetConfig() network.Config {
	cfg := network.DefaultConfig()
	cfg.DNSTimeout = 3 * time.Second
	cfg.DialTimeout = 3 * time.Second
	return cfg
}

// LoadScreen collects and formats one screen's body. It runs inside a
// tea.Cmd goroutine, never on the UI thread, and reuses the existing
// internal packages — no duplicated collection logic.
func LoadScreen(s Screen) string {
	switch s {
	case ScreenSystem:
		info, err := system.Collect()
		out := system.Format(info)
		if err != nil {
			out += "\nNote: some metrics unavailable.\n"
		}
		return out
	case ScreenProcesses:
		procs, err := process.Collect()
		if err != nil {
			return fmt.Sprintf("Could not list processes: %s\n", err)
		}
		process.Sort(procs, process.SortCPU)
		top, err := process.Top(procs, 15)
		if err != nil {
			return fmt.Sprintf("Could not list processes: %s\n", err)
		}
		return process.FormatTable(top)
	case ScreenDocker:
		containers, err := docker.List(false)
		if err != nil {
			return docker.UnavailableMessage()
		}
		return docker.FormatTable(containers)
	case ScreenGit:
		repo, err := git.Inspect("")
		if err != nil {
			return git.NotRepoMessage()
		}
		return git.Format(repo)
	case ScreenNetwork:
		return network.Format(network.CollectWith(quickNetConfig()))
	case ScreenDoctor:
		opts := doctor.DefaultOptions()
		opts.Net = quickNetConfig()
		return doctor.Format(doctor.RunWith(opts))
	default:
		return dashboardSummary()
	}
}

// dashboardSummary composes the overview from the same collectors
// the individual screens use.
func dashboardSummary() string {
	var b strings.Builder

	sys, _ := system.Collect()
	fmt.Fprintf(&b, "SYSTEM\n")
	if sys.HasCPU {
		fmt.Fprintf(&b, "  CPU      %.1f%% (%d cores)\n", sys.CPUPercent, sys.CPUCores)
	} else {
		b.WriteString("  CPU      N/A\n")
	}
	if sys.HasMem {
		fmt.Fprintf(&b, "  Memory   %.1f%%\n", sys.MemPercent)
	} else {
		b.WriteString("  Memory   N/A\n")
	}
	if sys.HasDisk {
		fmt.Fprintf(&b, "  Disk     %.1f%%\n", sys.DiskPercent)
	} else {
		b.WriteString("  Disk     N/A\n")
	}
	if sys.HasUptime {
		fmt.Fprintf(&b, "  Uptime   %s\n", shortUptime(sys.Uptime))
	}
	fmt.Fprintf(&b, "  OS       %s/%s\n", sys.OS, sys.Arch)

	b.WriteString("\nDEVELOPMENT\n")
	if repo, err := git.Inspect(""); err != nil {
		b.WriteString("  Git      (not a repository)\n")
	} else {
		state := "clean"
		if !repo.Clean {
			state = fmt.Sprintf("dirty (M%d S%d U%d)", len(repo.Modified), len(repo.Staged), len(repo.Untracked))
		}
		fmt.Fprintf(&b, "  Git      %s on %s\n", state, repo.Branch)
	}
	if containers, err := docker.List(false); err != nil {
		b.WriteString("  Docker   unavailable\n")
	} else {
		fmt.Fprintf(&b, "  Docker   %d container(s) running\n", len(containers))
	}

	b.WriteString("\nNETWORK\n")
	netInfo := network.CollectWith(quickNetConfig())
	fmt.Fprintf(&b, "  Internet %s\n", okMark(netInfo.NetOK))
	fmt.Fprintf(&b, "  DNS      %s\n", okMark(netInfo.DNSOK))
	if netInfo.HasLatency {
		fmt.Fprintf(&b, "  Latency  %s\n", netInfo.Latency.Round(time.Millisecond))
	} else {
		b.WriteString("  Latency  N/A\n")
	}

	b.WriteString("\nHEALTH\n")
	opts := doctor.DefaultOptions()
	opts.Net = quickNetConfig()
	sum := doctor.Summarize(doctor.RunWith(opts))
	fmt.Fprintf(&b, "  Passed %d  Warnings %d  Failed %d\n", sum.Passed, sum.Warnings, sum.Failed)

	return b.String()
}

func okMark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

func shortUptime(d time.Duration) string {
	days := int(d.Hours()) / 24
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, int(d.Hours())%24)
	}
	if d >= time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	if d >= time.Minute {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}
