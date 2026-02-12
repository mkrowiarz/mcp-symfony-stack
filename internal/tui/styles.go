package tui

import "github.com/charmbracelet/lipgloss"

var (
	nord0  = lipgloss.Color("#2E3440")
	nord1  = lipgloss.Color("#3B4252")
	nord3  = lipgloss.Color("#4C566A")
	nord4  = lipgloss.Color("#D8DEE9")
	nord8  = lipgloss.Color("#88C0D0")
	nord9  = lipgloss.Color("#81A1C1")
	nord10 = lipgloss.Color("#5E81AC")
	nord11 = lipgloss.Color("#BF616A")
	nord14 = lipgloss.Color("#A3BE8C")

	primaryColor   = nord8
	secondaryColor = nord3
	successColor   = nord14
	errorColor     = nord11

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			Padding(0, 1)

	paneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	focusedPaneStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	statusBarStyle = lipgloss.NewStyle().
			Background(nord10).
			Foreground(nord4).
			Padding(0, 1)

	modalStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(1, 3).
			Background(nord1)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(primaryColor).
			Padding(0, 1)

	confirmStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(errorColor).
			Padding(0, 1)
)
