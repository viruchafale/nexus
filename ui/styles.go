package ui

import (
	"github.com/charmbracelet/lipgloss"
)

// theme holds Lip Gloss styles. plain=true disables all color
// (NO_COLOR / NEXUS_NO_COLOR) while keeping layout intact.
type theme struct {
	title    lipgloss.Style
	subtitle lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	header   lipgloss.Style
	box      lipgloss.Style
	footer   lipgloss.Style
}

func newTheme(plain bool) theme {
	if plain {
		bold := lipgloss.NewStyle().Bold(true)
		return theme{
			title:    bold,
			subtitle: lipgloss.NewStyle(),
			selected: lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), false, false, false, true),
			normal:   lipgloss.NewStyle(),
			header:   bold,
			box:      lipgloss.NewStyle().Border(lipgloss.NormalBorder()).Padding(0, 1),
			footer:   lipgloss.NewStyle(),
		}
	}
	accent := lipgloss.Color("63") // violet-blue
	return theme{
		title:    lipgloss.NewStyle().Bold(true).Foreground(accent),
		subtitle: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		selected: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("229")).Background(lipgloss.Color("63")),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("252")),
		header:   lipgloss.NewStyle().Bold(true).Foreground(accent),
		box:      lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(accent).Padding(0, 1),
		footer:   lipgloss.NewStyle().Foreground(lipgloss.Color("240")),
	}
}
