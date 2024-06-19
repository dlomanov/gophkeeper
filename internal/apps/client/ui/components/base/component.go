package base

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Component interface {
	Title() string
	Init() InitResult
	Update(msg tea.Msg) UpdateResult
	View() string
}

type (
	InitResult struct {
		Status string
		Cmd    tea.Cmd
	}
	UpdateResult struct {
		Quitting     bool
		PassAccepted bool
		Next         Component
		Prev         Component
		Jump         Component
		Status       string
		Cmd          tea.Cmd
	}
	UpdateStatusMsg struct {
		Status string
	}
)

func (r UpdateResult) AppendCmd(cmds ...tea.Cmd) UpdateResult {
	r.Cmd = appendCmd(r.Cmd, cmds...)
	return r
}

func (r InitResult) AppendCmd(cmds ...tea.Cmd) InitResult {
	r.Cmd = appendCmd(r.Cmd, cmds...)
	return r
}

func appendCmd(cmd tea.Cmd, cmds ...tea.Cmd) tea.Cmd {
	if len(cmds) == 0 {
		return cmd
	}
	if cmd == nil && len(cmds) == 1 {
		return cmds[0]
	}
	batch := tea.Batch(cmds...)
	return tea.Batch(cmd, batch)
}

func UpdateStatusCmd(status string) tea.Cmd {
	return func() tea.Msg {
		return UpdateStatusMsg{Status: status}
	}
}
