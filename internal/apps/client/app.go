package client

import (
	"context"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/dlomanov/gophkeeper/internal/apps/client/config"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/client/migrations"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui"
	"github.com/dlomanov/gophkeeper/internal/apps/client/ui/components"
	"github.com/dlomanov/gophkeeper/internal/infra/logging"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"go.uber.org/zap"
)

type Model struct {
	tea.Model
	layout   *components.Layout
	curr     components.Component
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
		Level:       config.LogLevel,
		Type:        config.LogType,
		OutputPaths: config.LogOutputPaths,
	}); err != nil {
		return err
	}
	defer func(logger *zap.Logger) { _ = logger.Sync() }(logger)
	if c, err = deps.NewContainer(ctx, logger, config); err != nil {
		logger.Error("failed to init container", zap.Error(err))
		return err
	}
	defer closeContainer(c)
	if err = upMigrations(c); err != nil {
		return err
	}
	if err = c.Memcache.Load(ctx); err != nil {
		logger.Error("failed to load memcache", zap.Error(err))
		return err
	}

	runApp(c) // blocks current goroutine until termination

	return nil
}

func closeContainer(c *deps.Container) {
	if err := c.Close(); err != nil {
		c.Logger.Error("failed to close container", zap.Error(err))
	}
}

func upMigrations(c *deps.Container) error {
	ms, err := migrations.GetMigrations()
	if err != nil {
		c.Logger.Error("failed to get migrations", zap.Error(err))
		return err
	}
	if err = migrator.Migrate(c.Logger.Sugar(), c.DB.DB, ms); err != nil {
		c.Logger.Error("failed to up migrations", zap.Error(err))
		return err
	}
	return nil
}

func runApp(c *deps.Container) {
	c.Logger.Debug("starting app")
	model := ui.NewModel(c)
	if _, err := tea.NewProgram(model).Run(); err != nil {
		c.Logger.Error("app stopped with error", zap.Error(err))
		return
	}
	c.Logger.Debug("app stopped")
}
