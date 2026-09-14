package system

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/mem"
)

func stubProvider() provider {
	return provider{
		hostname:   func() (string, error) { return "testhost", nil },
		cpuInfo:    func() ([]cpu.InfoStat, error) { return []cpu.InfoStat{{ModelName: "Test CPU"}}, nil },
		cpuCounts:  func() (int, error) { return 8, nil },
		cpuPercent: func() ([]float64, error) { return []float64{12.5}, nil },
		virtMem: func() (*mem.VirtualMemoryStat, error) {
			return &mem.VirtualMemoryStat{Total: 16 << 30, Used: 8 << 30, UsedPercent: 50}, nil
		},
		diskUsage: func() (*disk.UsageStat, error) {
			return &disk.UsageStat{Total: 500 << 30, Used: 100 << 30, UsedPercent: 20}, nil
		},
		uptime: func() (uint64, error) { return 9000, nil },
	}
}

func TestCollectWithStubSuccess(t *testing.T) {
	info, err := collectWith(stubProvider())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Hostname != "testhost" || !info.HasHost {
		t.Errorf("hostname = %q, has=%v", info.Hostname, info.HasHost)
	}
	if info.CPUModel != "Test CPU" || info.CPUCores != 8 || !info.HasCPU {
		t.Errorf("cpu = %+v", info)
	}
	if info.MemTotal != 16<<30 || !info.HasMem {
		t.Errorf("mem = %+v", info)
	}
	if info.DiskTotal != 500<<30 || !info.HasDisk {
		t.Errorf("disk = %+v", info)
	}
	if info.Uptime != 9000*time.Second || !info.HasUptime {
		t.Errorf("uptime = %v, has=%v", info.Uptime, info.HasUptime)
	}
	if info.OS == "" || info.Arch == "" {
		t.Error("OS/Arch should always be set")
	}
}

func TestCollectWithPartialFailure(t *testing.T) {
	p := stubProvider()
	p.hostname = func() (string, error) { return "", errors.New("no hostname") }
	p.virtMem = func() (*mem.VirtualMemoryStat, error) { return nil, errors.New("no mem") }

	info, err := collectWith(p)
	if err == nil {
		t.Fatal("expected joined error for partial failure")
	}
	if info.HasHost {
		t.Error("HasHost should be false on hostname error")
	}
	if info.HasMem {
		t.Error("HasMem should be false on mem error")
	}
	// Unaffected metrics must still be present.
	if !info.HasCPU || !info.HasDisk || !info.HasUptime {
		t.Errorf("other metrics should survive partial failure: %+v", info)
	}
}

func TestFormatShowsAllFields(t *testing.T) {
	info, _ := collectWith(stubProvider())
	info.OS, info.Arch = "darwin", "arm64"
	out := Format(info)
	for _, want := range []string{
		"darwin", "arm64", "testhost",
		"Test CPU", "8", "12.5%",
		"16.0 GB", "8.0 GB", "50.0%",
		"500.0 GB", "100.0 GB", "20.0%",
		"2h 30m",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("Format missing %q\n%s", want, out)
		}
	}
}

func TestFormatNAOnEmpty(t *testing.T) {
	out := Format(Info{OS: "linux", Arch: "amd64"})
	if !strings.Contains(out, "N/A") {
		t.Errorf("expected N/A for missing metrics\n%s", out)
	}
	if !strings.Contains(out, "linux") {
		t.Errorf("OS should still render\n%s", out)
	}
}

func TestFormatBytes(t *testing.T) {
	cases := map[uint64]string{
		0:        "0 B",
		512:      "512 B",
		1024:     "1.0 KB",
		1536:     "1.5 KB",
		1 << 20:  "1.0 MB",
		1 << 30:  "1.0 GB",
		16 << 30: "16.0 GB",
		1 << 40:  "1.0 TB",
		5 << 40:  "5.0 TB",
	}
	for n, want := range cases {
		if got := formatBytes(n); got != want {
			t.Errorf("formatBytes(%d) = %q, want %q", n, got, want)
		}
	}
}

func TestFormatUptime(t *testing.T) {
	cases := map[time.Duration]string{
		45 * time.Second: "45s",
		90 * time.Second: "1m",
		90 * time.Minute: "1h 30m",
		3*24*time.Hour + 4*time.Hour + 12*time.Minute: "3d 4h 12m",
		2 * time.Hour: "2h 0m",
	}
	for d, want := range cases {
		if got := formatUptime(d); got != want {
			t.Errorf("formatUptime(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestPercentOrNA(t *testing.T) {
	if got := percentOrNA(12.345, true); got != "12.3%" {
		t.Errorf("got %q", got)
	}
	if got := percentOrNA(-5, true); got != "0.0%" {
		t.Errorf("negative should clamp, got %q", got)
	}
	if got := percentOrNA(150, true); got != "100.0%" {
		t.Errorf(">100 should clamp, got %q", got)
	}
	if got := percentOrNA(50, false); got != "N/A" {
		t.Errorf("unavailable should be N/A, got %q", got)
	}
}
