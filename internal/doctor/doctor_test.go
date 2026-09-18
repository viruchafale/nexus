package doctor

import (
	"fmt"
	"net"
	"strings"
	"testing"
	"time"

	"nexus/internal/network"
	"nexus/internal/system"
)

func TestUsageStatus(t *testing.T) {
	cases := []struct {
		pct        float64
		ok         bool
		warn, fail float64
		want       Status
	}{
		{10, true, 70, 90, Pass},
		{69.9, true, 70, 90, Pass},
		{70, true, 70, 90, Warn},
		{89.9, true, 70, 90, Warn},
		{90, true, 70, 90, Fail},
		{99, true, 70, 90, Fail},
		{0, false, 70, 90, Warn}, // unavailable is WARN, not FAIL
	}
	for _, c := range cases {
		if got := usageStatus(c.pct, c.ok, c.warn, c.fail); got != c.want {
			t.Errorf("usageStatus(%v,%v) = %v, want %v", c.pct, c.ok, got, c.want)
		}
	}
}

func TestCheckPortFreeAndOccupied(t *testing.T) {
	// Occupy an ephemeral port with a real listener.
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot listen on loopback, skipping port test")
	}
	defer ln.Close()
	port := ln.Addr().(*net.TCPAddr).Port

	if got := checkPort(port); got.Status != Warn {
		t.Errorf("occupied port should be WARN, got %+v", got)
	} else if !strings.Contains(got.Recommendation, fmt.Sprint(port)) {
		t.Errorf("recommendation should name the port: %+v", got)
	}

	free := checkPort(1) // port 1 is never a dev port; almost surely closed
	if free.Status != Pass {
		t.Logf("port 1 unexpectedly occupied in this environment: %+v", free)
	}
}

func TestSummarize(t *testing.T) {
	results := []Result{
		{Name: "a", Status: Pass}, {Name: "b", Status: Pass},
		{Name: "c", Status: Warn}, {Name: "d", Status: Fail},
	}
	sum := Summarize(results)
	if sum.Passed != 2 || sum.Warnings != 1 || sum.Failed != 1 {
		t.Errorf("got %+v", sum)
	}
}

func TestFormat(t *testing.T) {
	results := []Result{
		{Name: "CPU", Status: Pass, Detail: "usage 10.0%"},
		{Name: "Docker", Status: Warn, Detail: "daemon unavailable", Recommendation: "Start Docker if required"},
		{Name: "Network", Status: Fail, Detail: "no internet", Recommendation: "Check Wi-Fi"},
	}
	out := Format(results)
	for _, want := range []string{
		"NEXUS DOCTOR", "✓", "⚠", "✗",
		"CPU", "PASS", "Docker", "WARN", "Network", "FAIL",
		"SUMMARY", "Passed: 1", "Warnings: 1", "Failed: 1",
		"Warnings:", "- Docker: daemon unavailable", "- Network: no internet",
		"Recommendations:", "- Start Docker if required", "- Check Wi-Fi",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format missing %q\n%s", want, out)
		}
	}
}

func TestFormatAllPassOmitsSections(t *testing.T) {
	out := Format([]Result{{Name: "CPU", Status: Pass}})
	if strings.Contains(out, "\nWarnings:\n") || strings.Contains(out, "Recommendations:") {
		t.Errorf("all-pass report should omit warning sections\n%s", out)
	}
}

func TestResultConstructorsFromSnapshot(t *testing.T) {
	sys := system.Info{
		CPUPercent: 95, HasCPU: true,
		MemPercent: 50, HasMem: true,
		DiskPercent: 10, HasDisk: true,
	}
	if got := cpuResult(sys); got.Status != Fail {
		t.Errorf("cpu 95%% should FAIL, got %+v", got)
	}
	if got := memResult(sys); got.Status != Pass {
		t.Errorf("mem 50%% should PASS, got %+v", got)
	}
	if got := diskResult(sys); got.Status != Pass {
		t.Errorf("disk 10%% should PASS, got %+v", got)
	}
	if got := cpuResult(system.Info{}); got.Status != Warn || !strings.Contains(got.Detail, "unavailable") {
		t.Errorf("missing CPU metrics should WARN as unavailable, got %+v", got)
	}

	netInfo := network.NetInfo{NetOK: true, Latency: 20 * time.Millisecond, HasLatency: true, DNSOK: false, DNSDetail: "timeout"}
	if got := networkResult(netInfo); got.Status != Pass {
		t.Errorf("reachable net should PASS, got %+v", got)
	}
	if got := dnsResult(netInfo); got.Status != Fail {
		t.Errorf("failed DNS should FAIL, got %+v", got)
	}
}

func TestParsePorts(t *testing.T) {
	got, err := ParsePorts("3000,5432, 6379")
	if err != nil || len(got) != 3 || got[0] != 3000 || got[2] != 6379 {
		t.Errorf("got %v, err=%v", got, err)
	}
	got, err = ParsePorts("")
	if err != nil || len(got) != len(DevPorts) {
		t.Errorf("empty should mean DevPorts, got %v, err=%v", got, err)
	}
	for _, bad := range []string{"abc", "0", "70000", "3000,,5432", "-1"} {
		if _, err := ParsePorts(bad); err == nil {
			t.Errorf("ParsePorts(%q) should error", bad)
		}
	}
}

func TestRunWithCustomPorts(t *testing.T) {
	results := RunWith(Options{Ports: []int{5432}})
	names := map[string]bool{}
	for _, r := range results {
		names[r.Name] = true
	}
	if !names["Port 5432"] {
		t.Errorf("custom ports should be checked, got %v", names)
	}
	if names["Port 3000"] {
		t.Errorf("default ports should be skipped with custom list")
	}
	// Core checks always run regardless of port selection.
	for _, want := range []string{"CPU", "Memory", "Disk", "Network", "DNS", "Docker", "Git"} {
		if !names[want] {
			t.Errorf("missing core check %q", want)
		}
	}
}
