package tui

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	colorPrimary   = lipgloss.Color("205") // Pink
	colorUser      = lipgloss.Color("86")  // Cyan
	colorAssistant = lipgloss.Color("42")  // Green
	colorTool      = lipgloss.Color("226") // Yellow
	colorMuted     = lipgloss.Color("241") // Gray
	colorDanger    = lipgloss.Color("196") // Red

	// Header / Title styles
	styleHeader = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary).
			Padding(0, 1)

	styleBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("0")).
			Background(colorPrimary).
			Padding(0, 1)

	// Chat message styles
	styleUserMsg = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorUser)

	styleAssistantMsg = lipgloss.NewStyle().
				Foreground(colorAssistant)

	styleToolMsg = lipgloss.NewStyle().
			Foreground(colorTool).
			Italic(true)

	styleSystemMsg = lipgloss.NewStyle().
			Foreground(colorMuted).
			Italic(true)

	// Modal style
	styleModalBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorDanger).
			Padding(1, 2).
			Width(60)

	// Status bar style
	styleStatusBar = lipgloss.NewStyle().
			Foreground(lipgloss.Color("255")).
			Background(lipgloss.Color("237")).
			Padding(0, 1)

	// Menu active item highlight
	styleActiveTab = lipgloss.NewStyle().
			Bold(true).
			Foreground(colorPrimary)
)
