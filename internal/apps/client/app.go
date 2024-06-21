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
	if c, err = deps.NewContainer(logger, config); err != nil {
		logger.Error("failed to init container", zap.Error(err))
		return err
	}
	defer closeContainer(c)

	runApp(ctx, c)

	return nil
}

func closeContainer(c *deps.Container) {
	if err := c.Close(); err != nil {
		c.Logger.Error("failed to close container", zap.Error(err))
	}
}

func runApp(ctx context.Context, c *deps.Container) {
	c.Logger.Debug("starting app")
	model := ui.NewModel(c)
	if _, err := tea.NewProgram(model, tea.WithContext(ctx)).Run(); err != nil {
		c.Logger.Error("app stopped with error", zap.Error(err))
		return
	}
	c.Logger.Debug("app stopped")
}
