package process

import (
	"strings"
	"testing"
)

func sample() []Proc {
	return []Proc{
		{PID: 1, Name: "low", Status: "sleeping", CPUPercent: 1.0, MemPercent: 5.0, RSS: 100 << 20},
		{PID: 2, Name: "high-cpu", Status: "running", CPUPercent: 40.0, MemPercent: 2.0, RSS: 50 << 20},
		{PID: 3, Name: "high-mem", Status: "sleeping", CPUPercent: 2.0, MemPercent: 30.0, RSS: 2 << 30},
		{PID: 4, Name: "mid", Status: "running", CPUPercent: 10.0, MemPercent: 10.0, RSS: 500 << 20},
	}
}

func TestSortCPUDefault(t *testing.T) {
	procs := sample()
	Sort(procs, SortCPU)
	want := []int32{2, 4, 3, 1}
	for i, pid := range want {
		if procs[i].PID != pid {
			t.Fatalf("cpu order[%d] = pid %d, want %d (%v)", i, procs[i].PID, pid, procs)
		}
	}
}

func TestSortMem(t *testing.T) {
	procs := sample()
	Sort(procs, SortMem)
	want := []int32{3, 4, 1, 2}
	for i, pid := range want {
		if procs[i].PID != pid {
			t.Fatalf("mem order[%d] = pid %d, want %d (%v)", i, procs[i].PID, pid, procs)
		}
	}
}

func TestParseSortMode(t *testing.T) {
	for in, want := range map[string]SortMode{"cpu": SortCPU, "CPU": SortCPU, "mem": SortMem, "memory": SortMem, "MEMORY": SortMem} {
		got, err := ParseSortMode(in)
		if err != nil || got != want {
			t.Errorf("ParseSortMode(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"", "disk", "pid"} {
		if _, err := ParseSortMode(bad); err == nil {
			t.Errorf("ParseSortMode(%q) should error", bad)
		}
	}
}

func TestTop(t *testing.T) {
	procs := sample()
	top, err := Top(procs, 2)
	if err != nil || len(top) != 2 {
		t.Fatalf("Top(2) = %d items, err=%v", len(top), err)
	}
	if _, err := Top(procs, 0); err == nil {
		t.Error("Top(0) should error")
	}
	if _, err := Top(procs, -3); err == nil {
		t.Error("Top(-3) should error")
	}
	all, err := Top(procs, 100)
	if err != nil || len(all) != len(procs) {
		t.Errorf("Top(100) should return all, got %d, err=%v", len(all), err)
	}
}

func TestFormatTable(t *testing.T) {
	out := FormatTable(sample())
	for _, want := range []string{"PID", "CPU%", "MEM%", "MEMORY", "NAME", "STATUS", "high-cpu", "high-mem"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\n%s", want, out)
		}
	}
	if !strings.Contains(out, "2.0 GB") {
		t.Errorf("RSS should render human-readable\n%s", out)
	}
}

func TestFormatTableEmpty(t *testing.T) {
	out := FormatTable(nil)
	if !strings.Contains(out, "PID") || !strings.Contains(out, "no processes found") {
		t.Errorf("empty table should show header + note\n%s", out)
	}
}

func TestTruncate(t *testing.T) {
	if got := truncate("short", 32); got != "short" {
		t.Errorf("got %q", got)
	}
	long := strings.Repeat("x", 40)
	if got := truncate(long, 32); len([]rune(got)) > 32 {
		t.Errorf("not truncated: %q", got)
	}
}
