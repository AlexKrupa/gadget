package core

import "github.com/charmbracelet/lipgloss"

// Global styles for consistent UI appearance
var (
	// Main styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			PaddingTop(1).
			PaddingLeft(4).
			PaddingRight(4).
			PaddingBottom(1)

	StatusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#343433", Dark: "#C1C6B2"}).
			PaddingLeft(4).
			PaddingRight(4)

	// List styles
	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(4)

	SelectedItemStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#EE6FF8")).
				Bold(true).
				PaddingLeft(2).
				BorderLeft(true).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(lipgloss.Color("#EE6FF8"))

	// Text styles
	HelpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "#909090", Dark: "#626262"})

	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5F87")).
			Bold(true)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true)

	InfoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD"))

	// Input styles
	FocusedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EE6FF8"))

	BlurredStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262"))

	NoStyle = lipgloss.NewStyle()

	// Layout styles
	DocStyle = lipgloss.NewStyle().
			Padding(1, 2, 1, 2)
)
