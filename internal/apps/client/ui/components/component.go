package components

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	Title() string
	Init() tea.Cmd
	Update(msg tea.Msg) (UpdateResult, tea.Cmd)
	View() string
}

type UpdateResult struct {
	Quitting     bool
	PassAccepted bool
	Next         Component
	Prev         Component
	Jump         Component
	Status       string
}
