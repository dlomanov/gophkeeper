package styles

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
)

const (
	dotChar = " • "
)

var (
	SubtleStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	MainStyle    = lipgloss.NewStyle().MarginLeft(2)
	DotStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("236")).Render(dotChar)
	FocusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	BlurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	CursorStyle  = FocusedStyle
	NoStyle      = lipgloss.NewStyle()
	StatusStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	FocusedButton = FocusedStyle.Render("[ Submit ]")
	BlurredButton = fmt.Sprintf("[ %s ]", BlurredStyle.Render("Submit"))
)
