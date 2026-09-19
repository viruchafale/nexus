package ai

import (
	"fmt"
	"runtime"
	"strings"

	"nexus/internal/docker"
	"nexus/internal/doctor"
	"nexus/internal/git"
	"nexus/internal/network"
	"nexus/internal/system"
)

// CollectContext builds the read-only environment snapshot sent to the LLM.
// It reuses the existing collectors and contains machine facts only —
// no environment values, keys, file contents, or credentials of any kind.
func CollectContext() string {
	var b strings.Builder

	sys, _ := system.Collect()
	fmt.Fprintf(&b, "OS: %s/%s\n", runtime.GOOS, sys.Arch)
	if sys.HasCPU {
		fmt.Fprintf(&b, "CPU: %s, %d cores, usage %.1f%%\n", sys.CPUModel, sys.CPUCores, sys.CPUPercent)
	} else {
		b.WriteString("CPU: unavailable\n")
	}
	if sys.HasMem {
		fmt.Fprintf(&b, "Memory: usage %.1f%%\n", sys.MemPercent)
	} else {
		b.WriteString("Memory: unavailable\n")
	}
	if sys.HasDisk {
		fmt.Fprintf(&b, "Disk (/): usage %.1f%%\n", sys.DiskPercent)
	} else {
		b.WriteString("Disk: unavailable\n")
	}

	if containers, err := docker.List(false); err != nil {
		b.WriteString("Docker: unavailable\n")
	} else {
		fmt.Fprintf(&b, "Docker: %d container(s) running\n", len(containers))
		for _, c := range containers {
			fmt.Fprintf(&b, "  - %s (%s): %s\n", c.Name, c.Image, c.Status)
		}
	}

	if repo, err := git.Inspect(""); err != nil {
		b.WriteString("Git: not a repository\n")
	} else {
		state := "clean"
		if !repo.Clean {
			state = "dirty"
		}
		fmt.Fprintf(&b, "Git: %s on %s, latest %s %s\n",
			state, repo.Branch, repo.CommitHash, repo.CommitSubject)
	}

	netInfo := network.CollectWith(network.DefaultConfig())
	fmt.Fprintf(&b, "Network: internet=%s dns=%s\n", boolWord(netInfo.NetOK), boolWord(netInfo.DNSOK))

	sum := doctor.Summarize(doctor.Run())
	fmt.Fprintf(&b, "Doctor: %d passed, %d warnings, %d failed\n",
		sum.Passed, sum.Warnings, sum.Failed)

	return b.String()
}

func boolWord(ok bool) string {
	if ok {
		return "ok"
	}
	return "unreachable"
}
