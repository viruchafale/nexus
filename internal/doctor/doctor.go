package doctor

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"nexus/internal/docker"
	"nexus/internal/git"
	"nexus/internal/network"
	"nexus/internal/system"
)

// Status is the outcome of one check.
type Status int

const (
	Pass Status = iota
	Warn
	Fail
)

func (s Status) String() string {
	switch s {
	case Warn:
		return "WARN"
	case Fail:
		return "FAIL"
	default:
		return "PASS"
	}
}

// Result is the outcome of a single diagnostic check.
type Result struct {
	Name           string
	Status         Status
	Detail         string // one-line evidence, shown in Warnings on non-Pass
	Recommendation string // shown in Recommendations on non-Pass
}

// Check is one independent diagnostic. Checks are read-only and
// never fix, kill, restart, or modify anything.
type Check func() Result

// DevPorts are the default development ports probed for occupancy.
var DevPorts = []int{3000, 5173, 8000, 8080, 5432, 6379}

// portTimeout bounds each localhost probe.
const portTimeout = 500 * time.Millisecond

// Options tunes a doctor run. Zero value is valid: nil Ports means
// DevPorts, zero Net means network defaults (same 5s timeouts as
// `nexus network`, so both commands behave identically offline).
type Options struct {
	Ports []int
	Net   network.Config
}

// DefaultOptions returns the standard doctor configuration.
func DefaultOptions() Options {
	return Options{Ports: DevPorts, Net: network.DefaultConfig()}
}

// normalize fills unset fields with defaults.
func (o Options) normalize() Options {
	if o.Ports == nil {
		o.Ports = DevPorts
	}
	if o.Net.DNSTimeout <= 0 {
		o.Net.DNSTimeout = network.DefaultConfig().DNSTimeout
	}
	if o.Net.DialTimeout <= 0 {
		o.Net.DialTimeout = network.DefaultConfig().DialTimeout
	}
	if o.Net.DNSHost == "" {
		o.Net.DNSHost = network.DefaultConfig().DNSHost
	}
	if o.Net.DialAddr == "" {
		o.Net.DialAddr = network.DefaultConfig().DialAddr
	}
	return o
}

// Run executes all checks with default options.
func Run() []Result {
	return RunWith(DefaultOptions())
}

// RunWith executes all checks. The system and network snapshots are
// collected once and shared by the CPU/memory/disk and network/DNS
// checks respectively — results stay independent, collection doesn't repeat.
func RunWith(opts Options) []Result {
	opts = opts.normalize()
	sys, _ := system.Collect() // partial failures surface per-check via Has* flags
	netInfo := network.CollectWith(opts.Net)

	checks := []Check{
		func() Result { return cpuResult(sys) },
		func() Result { return memResult(sys) },
		func() Result { return diskResult(sys) },
		func() Result { return networkResult(netInfo) },
		func() Result { return dnsResult(netInfo) },
		checkDocker, checkGit,
	}
	results := make([]Result, 0, len(checks)+len(opts.Ports))
	for _, c := range checks {
		results = append(results, c())
	}
	for _, port := range opts.Ports {
		results = append(results, checkPort(port))
	}
	return results
}

// ParsePorts parses a comma-separated port list for --ports
// (e.g. "3000,5432,6379"). Empty string means DevPorts.
func ParsePorts(s string) ([]int, error) {
	if strings.TrimSpace(s) == "" {
		return DevPorts, nil
	}
	var ports []int
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(part)
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > 65535 {
			return nil, fmt.Errorf("invalid port %q: want 1-65535", part)
		}
		ports = append(ports, n)
	}
	return ports, nil
}

// Thresholds: below warn = PASS, warn..fail = WARN, above fail = FAIL.
const (
	warnCPU, failCPU   = 70.0, 90.0
	warnMem, failMem   = 80.0, 95.0
	warnDisk, failDisk = 80.0, 90.0
)

// usageStatus maps a percentage to a Status; ok=false means the
// metric was unavailable, which is a WARN (unknown, not failure).
func usageStatus(pct float64, ok bool, warnAt, failAt float64) Status {
	if !ok {
		return Warn
	}
	switch {
	case pct >= failAt:
		return Fail
	case pct >= warnAt:
		return Warn
	default:
		return Pass
	}
}

// cpuResult, memResult and diskResult are pure constructors over a shared
// system snapshot: independently testable, collected only once per run.
func cpuResult(info system.Info) Result {
	st := usageStatus(info.CPUPercent, info.HasCPU, warnCPU, failCPU)
	detail := fmt.Sprintf("usage %.1f%%", info.CPUPercent)
	if !info.HasCPU {
		detail = "CPU metrics unavailable"
	}
	return Result{Name: "CPU", Status: st, Detail: detail,
		Recommendation: "Close CPU-heavy processes (see: nexus processes)"}
}

func memResult(info system.Info) Result {
	st := usageStatus(info.MemPercent, info.HasMem, warnMem, failMem)
	detail := fmt.Sprintf("usage %.1f%%", info.MemPercent)
	if !info.HasMem {
		detail = "memory metrics unavailable"
	}
	return Result{Name: "Memory", Status: st, Detail: detail,
		Recommendation: "Close memory-heavy apps or add swap/RAM"}
}

func diskResult(info system.Info) Result {
	st := usageStatus(info.DiskPercent, info.HasDisk, warnDisk, failDisk)
	detail := fmt.Sprintf("usage %.1f%%", info.DiskPercent)
	if !info.HasDisk {
		detail = "disk metrics unavailable"
	}
	return Result{Name: "Disk", Status: st, Detail: detail,
		Recommendation: "Free disk space (caches, logs, unused images)"}
}

// networkResult and dnsResult read one shared network snapshot.
func networkResult(info network.NetInfo) Result {
	if info.NetOK {
		return Result{Name: "Network", Status: Pass,
			Detail: fmt.Sprintf("internet reachable (%s)", info.Latency.Round(time.Millisecond))}
	}
	return Result{Name: "Network", Status: Fail, Detail: "no internet: " + info.NetDetail,
		Recommendation: "Check Wi-Fi/cable, VPN, or firewall"}
}

func dnsResult(info network.NetInfo) Result {
	if info.DNSOK {
		return Result{Name: "DNS", Status: Pass, Detail: "example.com resolves"}
	}
	return Result{Name: "DNS", Status: Fail, Detail: "resolution failed: " + info.DNSDetail,
		Recommendation: "Check DNS servers (/etc/resolv.conf) or VPN"}
}

func checkDocker() Result {
	containers, err := docker.List(false)
	if err != nil {
		var unavail *docker.UnavailableError
		if errors.As(err, &unavail) {
			return Result{Name: "Docker", Status: Warn,
				Detail:         "Docker daemon unavailable",
				Recommendation: "Start Docker if required (Docker is optional)"}
		}
		return Result{Name: "Docker", Status: Warn, Detail: err.Error()}
	}
	return Result{Name: "Docker", Status: Pass,
		Detail: fmt.Sprintf("%d container(s) running", len(containers))}
}

func checkGit() Result {
	repo, err := git.Inspect("")
	if err != nil {
		var notRepo *git.NotRepoError
		if errors.As(err, &notRepo) {
			return Result{Name: "Git", Status: Warn,
				Detail:         "not a git repository",
				Recommendation: "Run inside a Git working tree for repo checks"}
		}
		return Result{Name: "Git", Status: Warn, Detail: err.Error()}
	}
	if repo.Clean {
		return Result{Name: "Git", Status: Pass,
			Detail: fmt.Sprintf("clean on %s", repo.Branch)}
	}
	return Result{Name: "Git", Status: Warn,
		Detail: fmt.Sprintf("dirty on %s (M%d S%d U%d)",
			repo.Branch, len(repo.Modified), len(repo.Staged), len(repo.Untracked)),
		Recommendation: "Review changes (see: nexus git)"}
}

// checkPort reports WARN when localhost:port accepts a connection
// (something already occupies it), PASS when free.
func checkPort(port int) Result {
	name := fmt.Sprintf("Port %d", port)
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("127.0.0.1:%d", port), portTimeout)
	if err != nil {
		return Result{Name: name, Status: Pass, Detail: "free"}
	}
	_ = conn.Close()
	return Result{Name: name, Status: Warn,
		Detail:         "already in use",
		Recommendation: fmt.Sprintf("Check the process using port %d", port)}
}
