package system

import (
	"errors"
	"os"
	"runtime"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

// Info holds a snapshot of system information.
// Has* flags mark metrics that were unavailable instead of
// inventing fake values.
type Info struct {
	OS       string
	Arch     string
	Hostname string
	HasHost  bool

	CPUModel   string
	CPUCores   int
	CPUPercent float64
	HasCPU     bool

	MemTotal   uint64
	MemUsed    uint64
	MemPercent float64
	HasMem     bool

	DiskTotal   uint64
	DiskUsed    uint64
	DiskPercent float64
	HasDisk     bool

	Uptime    time.Duration
	HasUptime bool
}

// provider groups external calls so collectWith is testable
// without touching the real machine.
type provider struct {
	hostname   func() (string, error)
	cpuInfo    func() ([]cpu.InfoStat, error)
	cpuCounts  func() (int, error)
	cpuPercent func() ([]float64, error)
	virtMem    func() (*mem.VirtualMemoryStat, error)
	diskUsage  func() (*disk.UsageStat, error)
	uptime     func() (uint64, error)
}

func defaultProvider() provider {
	return provider{
		hostname:   os.Hostname,
		cpuInfo:    cpu.Info,
		cpuCounts:  func() (int, error) { return cpu.Counts(true) },
		cpuPercent: func() ([]float64, error) { return cpu.Percent(500*time.Millisecond, false) },
		virtMem:    mem.VirtualMemory,
		diskUsage:  func() (*disk.UsageStat, error) { return disk.Usage("/") },
		uptime:     host.Uptime,
	}
}

// Collect gathers system info, tolerating partial failures.
// Returned error joins per-metric failures; Info still holds
// whatever was available (check Has* flags).
func Collect() (Info, error) {
	return collectWith(defaultProvider())
}

func collectWith(p provider) (Info, error) {
	var errs []error
	info := Info{OS: runtime.GOOS, Arch: runtime.GOARCH}

	if h, err := p.hostname(); err != nil {
		errs = append(errs, err)
	} else {
		info.Hostname, info.HasHost = h, true
	}

	cpuModel, cores, pct, cpuOK := collectCPU(p, &errs)
	info.CPUModel, info.CPUCores, info.CPUPercent, info.HasCPU = cpuModel, cores, pct, cpuOK

	if vm, err := p.virtMem(); err != nil {
		errs = append(errs, err)
	} else if vm == nil {
		errs = append(errs, errors.New("system: nil memory stats"))
	} else {
		info.MemTotal, info.MemUsed, info.MemPercent, info.HasMem = vm.Total, vm.Used, vm.UsedPercent, true
	}

	if du, err := p.diskUsage(); err != nil {
		errs = append(errs, err)
	} else if du == nil {
		errs = append(errs, errors.New("system: nil disk stats"))
	} else {
		info.DiskTotal, info.DiskUsed, info.DiskPercent, info.HasDisk = du.Total, du.Used, du.UsedPercent, true
	}

	if up, err := p.uptime(); err != nil {
		errs = append(errs, err)
	} else {
		info.Uptime, info.HasUptime = time.Duration(up)*time.Second, true
	}

	return info, errors.Join(errs...)
}

func collectCPU(p provider, errs *[]error) (string, int, float64, bool) {
	model, cores, pct := "", 0, 0.0
	ok := true

	infos, err := p.cpuInfo()
	if err != nil {
		*errs = append(*errs, err)
		ok = false
	} else if len(infos) > 0 {
		model = infos[0].ModelName
	}

	n, err := p.cpuCounts()
	if err != nil {
		*errs = append(*errs, err)
		ok = false
	} else {
		cores = n
	}

	percents, err := p.cpuPercent()
	if err != nil {
		*errs = append(*errs, err)
		ok = false
	} else if len(percents) > 0 {
		pct = percents[0]
	}

	if cores <= 0 && model == "" {
		ok = false
	}
	return model, cores, pct, ok
}
