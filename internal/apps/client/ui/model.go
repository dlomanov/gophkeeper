package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components"
)

type Model struct {
	tea.Model
	layout   *components.Layout
	curr     components.Component
	quitting bool
	status   string
}

func NewModel(c *deps.Container) Model {
	main := components.NewMain("gophkeeper", nil)
	table := components.NewEntryTable("gophkeeper/entries", nil, c.EntryUC, c.Logger)
	settings := components.NewSettings("gophkeeper/settings", nil)
	signUp := components.NewSignUp("gophkeeper/sync/sign-up", nil, c.UserUC, c.Memcache)
	signIn := components.NewSignIn("gophkeeper/sync/sign-in", nil, c.UserUC, c.Memcache)
	menu := components.NewMenu("gophkeeper", main, []components.Nav{
		{Name: "Sign-up", Next: signUp},
		{Name: "Sign-in", Next: signIn},
		{Name: "Entries", Next: table},
		{Name: "Settings", Next: settings},
	})
	table.SetPrev(menu)
	settings.SetPrev(menu)
	signUp.SetPrev(menu)
	signIn.SetPrev(menu)
	main.SetNext(menu)
	return Model{
		layout: components.NewLayout(),
		curr:   main,
	}
}

func (m Model) Init() tea.Cmd {
	return m.curr.Init()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg, ok := msg.(tea.KeyMsg); ok {
		if k := msg.String(); k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	}
	res, cmd := m.curr.Update(msg)
	switch {
	case res.Next != nil:
		m.curr = res.Next
		m.status = ""
		cmd = m.curr.Init()
	case res.Prev != nil:
		m.curr = res.Prev
		m.status = ""
		cmd = m.curr.Init()
	case res.Jump != nil:
		m.curr = res.Jump
		m.status = ""
		cmd = m.curr.Init()
	}
	if res.Status != "" {
		m.status = res.Status
	}
	m.quitting = res.Quitting

	return m, cmd
}

func (m Model) View() string {
	return m.layout.View(m.curr, m.quitting, m.status)
}
