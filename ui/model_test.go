package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func keyMsg(s string) tea.KeyMsg {
	// Build a KeyMsg the same way the terminal driver does.
	var k tea.KeyMsg
	switch s {
	case "up":
		k = tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		k = tea.KeyMsg{Type: tea.KeyDown}
	case "enter":
		k = tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		k = tea.KeyMsg{Type: tea.KeyEsc}
	default:
		k = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(s)}
	}
	return k
}

func TestNavigateClamps(t *testing.T) {
	m := New(true)
	updated, _ := m.Update(keyMsg("up"))
	m = updated.(Model)
	if m.cursor != ScreenDashboard {
		t.Errorf("up at top should clamp, got %v", m.cursor)
	}
	for i := 0; i < 20; i++ {
		updated, _ := m.Update(keyMsg("down"))
		m = updated.(Model)
	}
	if m.cursor != ScreenDoctor {
		t.Errorf("down should clamp at last screen, got %v", m.cursor)
	}
}

func TestEnterOpensAndLoads(t *testing.T) {
	m := New(true)
	updated, cmd := m.Update(keyMsg("down")) // cursor -> System
	m = updated.(Model)
	updated, cmd = m.Update(keyMsg("enter"))
	m = updated.(Model)
	if m.active != ScreenSystem {
		t.Errorf("enter should open cursor screen, got %v", m.active)
	}
	if cmd == nil {
		t.Error("opening an unloaded screen should start a load")
	}
}

func TestEnterLoadedScreenNoReload(t *testing.T) {
	m := New(true)
	m.bodies[ScreenGit] = "cached"
	m.cursor = ScreenGit
	updated, cmd := m.Update(keyMsg("enter"))
	m = updated.(Model)
	if m.active != ScreenGit || cmd != nil {
		t.Errorf("opening a loaded screen should not reload (cmd=%v)", cmd)
	}
}

func TestDataMsgStoresBody(t *testing.T) {
	m := New(true)
	m.loading[ScreenGit] = true
	updated, _ := m.Update(dataMsg{screen: ScreenGit, body: "hello"})
	m = updated.(Model)
	if m.bodies[ScreenGit] != "hello" || m.loading[ScreenGit] {
		t.Errorf("dataMsg should store body and clear loading: %+v", m)
	}
}

func TestRefreshReloads(t *testing.T) {
	m := New(true)
	m.active = ScreenDocker
	m.bodies[ScreenDocker] = "old"
	updated, cmd := m.Update(keyMsg("r"))
	m = updated.(Model)
	if cmd == nil {
		t.Error("r should trigger a reload")
	}
	if _, ok := m.bodies[ScreenDocker]; ok {
		t.Error("r should drop the stale body")
	}
}

func TestEscReturnsHome(t *testing.T) {
	m := New(true)
	m.active, m.cursor = ScreenNetwork, ScreenNetwork
	updated, _ := m.Update(keyMsg("esc"))
	m = updated.(Model)
	if m.active != ScreenDashboard || m.cursor != ScreenDashboard {
		t.Errorf("esc should return home, got active=%v cursor=%v", m.active, m.cursor)
	}
}

func TestQuit(t *testing.T) {
	m := New(true)
	_, cmd := m.Update(keyMsg("q"))
	if cmd == nil {
		t.Error("q should quit")
	}
}

func TestViewContainsChrome(t *testing.T) {
	m := New(true)
	m.width = 100
	out := m.View()
	for _, want := range []string{"NEXUS", "Developer Command Center", "Dashboard", "System", "Doctor", "↑/↓", "q quit"} {
		if !strings.Contains(out, want) {
			t.Errorf("view missing %q\n%s", want, out)
		}
	}
}

func TestTitles(t *testing.T) {
	for s := ScreenDashboard; s < ScreenCount; s++ {
		if s.Title() == "" {
			t.Errorf("screen %d missing title", s)
		}
	}
}
