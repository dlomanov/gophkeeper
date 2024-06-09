package ui

import (
	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	Title() string
	Init() tea.Cmd
	Update(msg tea.Msg) (UpdateResult, tea.Cmd)
	View() string
}

type UpdateResult struct {
	Quitting bool
	Next     Component
	Prev     Component
	Jump     Component
	Status   string
}

func passwordInput(placeholder string) textinput.Model {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.CharLimit = 32
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Cursor.SetMode(cursor.CursorBlink)
	return ti
}
