package components

import (
	"github.com/charmbracelet/lipgloss"
)

const (
	dotChar = " • "
)

var (
	keywordStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("211"))
	subtleStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	ticksStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("79"))
	checkboxStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("212"))
	mainStyle     = lipgloss.NewStyle().MarginLeft(2)
	dotStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)

	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	errStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	cursorStyle  = focusedStyle

	noStyle     = lipgloss.NewStyle()
	helpStyle   = blurredStyle
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))
)
