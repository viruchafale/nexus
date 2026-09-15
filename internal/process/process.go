package process

import (
	"sort"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/process"
)

// Proc is a read-only snapshot of one process.
// Zero CPU/Mem values mean "unavailable", never kill/modify.
type Proc struct {
	PID        int32
	Name       string
	Status     string
	CPUPercent float64
	MemPercent float64
	RSS        uint64
}

// SortMode selects the ranking for Top.
type SortMode string

const (
	SortCPU SortMode = "cpu"
	SortMem SortMode = "mem"
)

// ParseSortMode validates the --sort flag value.
func ParseSortMode(s string) (SortMode, error) {
	switch SortMode(strings.ToLower(strings.TrimSpace(s))) {
	case SortCPU:
		return SortCPU, nil
	case SortMem, "memory":
		return SortMem, nil
	default:
		return "", &SortError{Value: s}
	}
}

// SortError is returned for an invalid --sort value.
type SortError struct{ Value string }

func (e *SortError) Error() string {
	return "invalid --sort value " + quote(e.Value) + ": want \"cpu\" or \"memory\""
}

func quote(s string) string {
	if s == "" {
		return "(empty)"
	}
	return "\"" + s + "\""
}

// Collect snapshots running processes. It never kills or modifies anything.
// Processes that vanish mid-collection or deny inspection are skipped and
// counted; only a failure to list processes at all is returned as an error.
func Collect() ([]Proc, error) {
	procs, err := process.Processes()
	if err != nil {
		return nil, err
	}
	// Prime per-process CPU counters, wait briefly, then read.
	// A single sample without this reads 0 for most processes.
	for _, p := range procs {
		_, _ = p.CPUPercent()
	}
	time.Sleep(200 * time.Millisecond)

	out := make([]Proc, 0, len(procs))
	for _, p := range procs {
		pr, ok := snapshot(p)
		if !ok {
			continue
		}
		out = append(out, pr)
	}
	return out, nil
}

// snapshot reads one process; ok=false means it vanished or is
// unreadable — the caller skips it instead of crashing.
func snapshot(p *process.Process) (Proc, bool) {
	name, err := p.Name()
	if err != nil || name == "" {
		return Proc{}, false
	}
	pr := Proc{PID: p.Pid, Name: name}
	if pct, err := p.CPUPercent(); err == nil {
		pr.CPUPercent = pct
	}
	if pct, err := p.MemoryPercent(); err == nil {
		pr.MemPercent = float64(pct)
	}
	if mi, err := p.MemoryInfo(); err == nil && mi != nil {
		pr.RSS = mi.RSS
	}
	if st, err := p.Status(); err == nil && len(st) > 0 {
		pr.Status = strings.Join(st, ",")
	} else {
		pr.Status = "-"
	}
	return pr, true
}

// Sort orders procs in place, highest first.
func Sort(procs []Proc, mode SortMode) {
	switch mode {
	case SortMem:
		sort.Slice(procs, func(i, j int) bool {
			if procs[i].MemPercent == procs[j].MemPercent {
				return procs[i].RSS > procs[j].RSS
			}
			return procs[i].MemPercent > procs[j].MemPercent
		})
	default:
		sort.Slice(procs, func(i, j int) bool {
			if procs[i].CPUPercent == procs[j].CPUPercent {
				return procs[i].MemPercent > procs[j].MemPercent
			}
			return procs[i].CPUPercent > procs[j].CPUPercent
		})
	}
}

// Top returns at most limit entries. limit < 1 returns an error.
func Top(procs []Proc, limit int) ([]Proc, error) {
	if limit < 1 {
		return nil, &LimitError{Value: limit}
	}
	if len(procs) > limit {
		return procs[:limit], nil
	}
	return procs, nil
}

// LimitError is returned for an invalid --limit value.
type LimitError struct{ Value int }

func (e *LimitError) Error() string {
	return "invalid --limit: must be >= 1"
}
