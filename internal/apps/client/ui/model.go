package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components"
	"strings"
)

type (
	Model struct {
		tea.Model
		c        *deps.Container
		curr     components.Component
		quitting bool
		accepted bool
		status   string
	}
)

func NewModel(c *deps.Container) Model {
	main := components.NewMain("gophkeeper", c)
	return Model{
		c:    c,
		curr: main,
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
	if res.PassAccepted && !m.accepted {
		m.accepted = true
		table := components.NewEntryTable("gophkeeper/entries", nil, m.c.EntryUC, m.c.Logger)
		settings := components.NewSettings("gophkeeper/settings", nil)
		signUp := components.NewSignUp("gophkeeper/sync/sign-up", nil, m.c.UserUC, m.c.Memcache)
		signIn := components.NewSignIn("gophkeeper/sync/sign-in", nil, m.c.UserUC, m.c.Memcache)
		menu := components.NewMenu("gophkeeper", []components.Nav{
			{Name: "Sign-up", Next: signUp},
			{Name: "Sign-in", Next: signIn},
			{Name: "Entries", Next: table},
			{Name: "Settings", Next: settings},
		})
		table.SetPrev(menu)
		settings.SetPrev(menu)
		signUp.SetPrev(menu)
		signIn.SetPrev(menu)
		m.curr = menu
		return m, m.curr.Init()
	}
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
	sb := strings.Builder{}
	sb.WriteByte('\n')
	sb.WriteString(m.curr.Title())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(m.curr.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if m.quitting {
		sb.WriteString(components.StatusStyle.Render("see you later! 😊\n\n"))
	} else {
		sb.WriteString(components.StatusStyle.Render(m.status))
	}

	return components.MainStyle.Render(sb.String())
}
