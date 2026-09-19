package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

// Run launches the interactive dashboard. It must be called from a
// real terminal; otherwise Bubble Tea returns an error instead of
// crashing, and the caller prints a hint.
func Run(noColor bool) error {
	p := tea.NewProgram(New(noColor))
	if _, err := p.Run(); err != nil {
		return fmt.Errorf("dashboard needs an interactive terminal: %w", err)
	}
	return nil
}
