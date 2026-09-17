package network

import (
	"context"
	"net"
	"os"
	"time"
)

// Config tunes timeouts and targets. Use DefaultConfig in production;
// tests and callers can shrink timeouts for offline environments.
type Config struct {
	DNSTimeout  time.Duration
	DialTimeout time.Duration
	// DNSHost is resolved to test DNS; DialAddr (IP literal, no DNS
	// needed) is dialed to test internet connectivity separately.
	DNSHost  string
	DialAddr string
}

// DefaultConfig returns MVP timeouts and well-known stable targets.
func DefaultConfig() Config {
	return Config{
		DNSTimeout:  5 * time.Second,
		DialTimeout: 5 * time.Second,
		DNSHost:     "example.com",
		DialAddr:    "1.1.1.1:443",
	}
}

// IfaceInfo is one network interface with its addresses.
type IfaceInfo struct {
	Name string
	IPs  []string
	Up   bool
}

// NetInfo is a read-only diagnostics snapshot.
// DNSOK/NetOK false simply means unreachable — never an error,
// so offline machines still get a full report.
type NetInfo struct {
	Hostname string
	HasHost  bool

	Interfaces []IfaceInfo

	DNSOK     bool
	DNSDetail string

	NetOK      bool
	NetDetail  string
	Latency    time.Duration
	HasLatency bool
}

// Collect runs diagnostics with default timeouts.
func Collect() NetInfo {
	return CollectWith(DefaultConfig())
}

// CollectWith runs diagnostics. It performs no scans, opens no listeners,
// and modifies nothing: one DNS lookup plus one TCP dial at most.
func CollectWith(cfg Config) NetInfo {
	var info NetInfo

	if h, err := os.Hostname(); err == nil && h != "" {
		info.Hostname, info.HasHost = h, true
	}
	info.Interfaces = listInterfaces()

	if ips, err := lookupHost(cfg); err != nil {
		info.DNSDetail = shortErr(err)
	} else {
		info.DNSOK = true
		info.DNSDetail = firstIP(ips)
	}

	if latency, err := dialLatency(cfg); err != nil {
		info.NetDetail = shortErr(err)
	} else {
		info.NetOK = true
		info.Latency, info.HasLatency = latency, true
	}

	return info
}

// listInterfaces returns non-loopback interfaces and their addresses.
// Loopback is skipped (uninteresting); down interfaces are listed
// without addresses so the user sees they exist.
func listInterfaces() []IfaceInfo {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil
	}
	var out []IfaceInfo
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		info := IfaceInfo{Name: iface.Name, Up: iface.Flags&net.FlagUp != 0}
		if !info.Up {
			out = append(out, info)
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			out = append(out, info)
			continue
		}
		for _, a := range addrs {
			if ip := addrIP(a); ip != "" {
				info.IPs = append(info.IPs, ip)
			}
		}
		out = append(out, info)
	}
	return out
}

// addrIP extracts the host IP from a net.Addr, skipping link-local IPv6
// (fe80::/10) noise that clutters the display.
func addrIP(a net.Addr) string {
	var ip net.IP
	switch v := a.(type) {
	case *net.IPNet:
		ip = v.IP
	case *net.IPAddr:
		ip = v.IP
	default:
		return ""
	}
	if ip == nil || ip.IsLoopback() || ip.IsLinkLocalUnicast() {
		return ""
	}
	return ip.String()
}

func lookupHost(cfg Config) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DNSTimeout)
	defer cancel()
	return net.DefaultResolver.LookupHost(ctx, cfg.DNSHost)
}

func dialLatency(cfg Config) (time.Duration, error) {
	d := net.Dialer{Timeout: cfg.DialTimeout}
	start := time.Now()
	conn, err := d.DialContext(context.Background(), "tcp", cfg.DialAddr)
	if err != nil {
		return 0, err
	}
	_ = conn.Close()
	return time.Since(start), nil
}

func firstIP(ips []string) string {
	if len(ips) == 0 {
		return "resolved"
	}
	return ips[0]
}

func shortErr(err error) string {
	if err == nil {
		return ""
	}
	if s := err.Error(); len(s) > 120 {
		return s[:117] + "..."
	} else {
		return s
	}
}
