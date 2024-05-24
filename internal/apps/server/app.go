package server

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/infra/logging"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"go.uber.org/zap"
)

func Run(config *config.Config) error {
	logger, err := logging.NewLogger(logging.Config{
		Level: config.LogLevel,
		Type:  config.LogType,
	})
	if err != nil {
		return err
	}

	c, err := deps.NewContainer(logger, config)
	if err != nil {
		return err
	}
	defer closeContainer(c)

	if err = upMigrations(c); err != nil {
		return err
	}

	return nil
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

func closeContainer(c *deps.Container) {
	if err := c.Close(); err != nil {
		c.Logger.Error("failed to close container", zap.Error(err))
	}
}
