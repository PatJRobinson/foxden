package app

import "github.com/charmbracelet/lipgloss"

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Underline(true).
			Padding(0, 1)

	inactiveTabStyle = lipgloss.NewStyle().
				Padding(0, 1)

	selectedStoryStyle = lipgloss.NewStyle().
				Bold(true)

	mutedStyle = lipgloss.NewStyle().
			Faint(true)

	footerStyle = lipgloss.NewStyle().
			Faint(true)

	boxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)
	overlayStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			Padding(1, 2)

	errorStyle = lipgloss.NewStyle().
			Bold(true)
)
