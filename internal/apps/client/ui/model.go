package ui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/navlist"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components/base/styles"
	"strings"
)

type (
	Model struct {
		tea.Model
		c        *deps.Container
		curr     base.Component
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
	result := m.curr.Init()
	return tea.Batch(tea.DisableMouse, result.Cmd)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if k := msg.String(); k == "ctrl+c" {
			m.quitting = true
			return m, tea.Quit
		}
	case base.UpdateStatusMsg:
		m.status = msg.Status
		return m, nil
	}
	var cmds tea.Cmd
	res := m.curr.Update(msg)
	cmds = tea.Batch(cmds, res.Cmd)
	if res.PassAccepted && !m.accepted {
		m.accepted = true
		table := components.NewEntryTable("gophkeeper/entries", m.c.EntryUC, m.c.Logger)
		signUp := components.NewSignUp("gophkeeper/sync/sign-up", m.c.UserUC, m.c.Memcache)
		signIn := components.NewSignIn("gophkeeper/sync/sign-in", m.c.UserUC, m.c.Memcache)
		about := components.NewSettings("gophkeeper/about", components.BuildInfo{
			Version: m.c.Config.BuildVersion,
			Date:    m.c.Config.BuildDate,
			Commit:  m.c.Config.BuildCommit,
		})
		menu := components.NewMenu("gophkeeper", []navlist.Item{
			{Name: "Sign-up", Next: signUp},
			{Name: "Sign-in", Next: signIn},
			{Name: "Entries", Next: table},
			{Name: "About", Next: about},
		})
		table.SetPrev(menu)
		signUp.SetPrev(menu)
		signIn.SetPrev(menu)
		about.SetPrev(menu)
		m.curr = menu
		result := m.curr.Init()
		m.status = result.Status
		return m, result.Cmd
	}
	if res.Status != "" {
		m.status = res.Status
	}
	switch {
	case res.Next != nil:
		m.curr = res.Next
		result := m.curr.Init()
		m.status = result.Status
		cmds = tea.Batch(cmds, result.Cmd)
	case res.Prev != nil:
		m.curr = res.Prev
		result := m.curr.Init()
		m.status = result.Status
		cmds = tea.Batch(cmds, result.Cmd)
	case res.Jump != nil:
		m.curr = res.Jump
		result := m.curr.Init()
		m.status = result.Status
		cmds = tea.Batch(cmds, result.Cmd)
	}
	m.quitting = res.Quitting

	return m, cmds
}

func (m Model) View() string {
	sb := strings.Builder{}
	sb.WriteByte('\n')
	sb.WriteString(m.curr.Title())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteString(m.curr.View())
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	if m.quitting {
		sb.WriteString(styles.StatusStyle.Render("see you later! 😊"))
	} else {
		sb.WriteString(styles.StatusStyle.Render(m.status))
	}
	sb.WriteByte('\n')
	sb.WriteByte('\n')
	return styles.MainStyle.Render(sb.String())
}
