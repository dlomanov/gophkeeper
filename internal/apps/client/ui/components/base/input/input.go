package input

import tea "github.com/charmbracelet/bubbletea"

type Input interface {
	Value() string
	SetValue(v string)
	Update(msg tea.Msg) tea.Cmd
	View() string
	Focus() tea.Cmd
	Blur()
	Reset()
}
