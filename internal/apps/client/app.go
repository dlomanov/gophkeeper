package client

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui"
	"github.com/dlomanov/gophkeeper/internal/infra/logging"
	"go.uber.org/zap"
)

type Model struct {
	tea.Model
	layout   *ui.Layout
	curr     ui.Component
	quitting bool
	status   string
}

func Run(ctx context.Context, config *config.Config) error {
	if err := config.Validate(); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}
	var (
		logger *zap.Logger
		c      *deps.Container
		err    error
	)
	if logger, err = logging.NewLogger(logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	}); err != nil {
		return err
	}
	defer func(logger *zap.Logger) { _ = logger.Sync() }(logger)
	if c, err = deps.NewContainer(ctx, logger, config); err != nil {
		logger.Error("failed to init container", zap.Error(err))
		return err
	}
	defer closeContainer(c)

	model := newModel(c)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		logger.Error("program stopped with error", zap.Error(err))
	}

	return nil
}

func closeContainer(c *deps.Container) {
	if err := c.Close(); err != nil {
		c.Logger.Error("failed to close container", zap.Error(err))
	}
}

func newModel(c *deps.Container) Model {
	main := ui.NewMain("gophkeeper", nil)
	table := ui.NewTable("gophkeeper/entries", nil, c.EntryUC)
	settings := ui.NewSettings("gophkeeper/settings", nil)
	signUp := ui.NewSignUp("gophkeeper/sync/sign-up", nil, c.UserUC, c.Cache)
	signIn := ui.NewSignIn("gophkeeper/sync/sign-in", nil, c.UserUC, c.Cache)
	menu := ui.NewMenu("gophkeeper", main, []ui.Nav{
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
		layout: ui.NewLayout(),
		curr:   main,
	}
}

func (m Model) Init() tea.Cmd {
	// TODO: possible spot for start background
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
		cmd = m.curr.Init()
	case res.Prev != nil:
		m.curr = res.Prev
		cmd = m.curr.Init()
	case res.Jump != nil:
		m.curr = res.Jump
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
