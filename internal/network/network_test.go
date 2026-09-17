package network

import (
	"strings"
	"testing"
	"time"
)

func onlineSample() NetInfo {
	return NetInfo{
		Hostname: "mac.local", HasHost: true,
		Interfaces: []IfaceInfo{
			{Name: "en0", Up: true, IPs: []string{"192.168.1.5"}},
			{Name: "awdl0", Up: false},
		},
		DNSOK: true, DNSDetail: "93.184.216.34",
		NetOK: true, Latency: 24 * time.Millisecond, HasLatency: true,
	}
}

func TestFormatOnline(t *testing.T) {
	out := Format(onlineSample())
	for _, want := range []string{"NETWORK", "INTERFACES", "✓", "24 ms", "en0", "192.168.1.5", "awdl0", "(down)", "mac.local"} {
		if !strings.Contains(out, want) {
			t.Errorf("report missing %q\n%s", want, out)
		}
	}
}

func TestFormatOffline(t *testing.T) {
	info := NetInfo{
		Interfaces: []IfaceInfo{{Name: "en0", Up: true}},
		DNSDetail:  "no such host",
		NetDetail:  "connection refused",
	}
	out := Format(info)
	for _, want := range []string{"✗", "N/A", "INTERFACES", "connection refused"} {
		if !strings.Contains(out, want) {
			t.Errorf("offline report missing %q\n%s", want, out)
		}
	}
	// Offline machines still get a full report, never a crash/blank.
	if !strings.Contains(out, "NETWORK") {
		t.Errorf("offline report missing NETWORK section\n%s", out)
	}
}

func TestFormatNoInterfaces(t *testing.T) {
	out := Format(NetInfo{})
	if !strings.Contains(out, "no interfaces found") {
		t.Errorf("got\n%s", out)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := map[time.Duration]string{
		24 * time.Millisecond:   "24 ms",
		1500 * time.Millisecond: "1.5 s",
		0:                       "0 ms",
		-5 * time.Millisecond:   "0 ms",
	}
	for d, want := range cases {
		if got := formatDuration(d); got != want {
			t.Errorf("formatDuration(%v) = %q, want %q", d, got, want)
		}
	}
}

func TestFormatIPs(t *testing.T) {
	if got := formatIPs(IfaceInfo{Name: "en0", Up: true, IPs: []string{"10.0.0.2", "2001:db8::1"}}); got != "10.0.0.2, 2001:db8::1" {
		t.Errorf("got %q", got)
	}
	if got := formatIPs(IfaceInfo{Name: "x", Up: false}); got != "(down)" {
		t.Errorf("got %q", got)
	}
	if got := formatIPs(IfaceInfo{Name: "x", Up: true}); got != "-" {
		t.Errorf("got %q", got)
	}
}

func TestCollectLoopbackLocal(t *testing.T) {
	// Local-only: hostile-network-proof assertion — interface listing
	// must exclude loopback and never error.
	ifaces := listInterfaces()
	for _, iface := range ifaces {
		if iface.Name == "lo" || iface.Name == "lo0" {
			t.Errorf("loopback should be skipped, got %+v", iface)
		}
		for _, ip := range iface.IPs {
			if ip == "127.0.0.1" || ip == "::1" {
				t.Errorf("loopback IP leaked into %s: %v", iface.Name, iface.IPs)
			}
		}
	}
}
