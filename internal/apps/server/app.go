package server

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/server/migrations"
	"github.com/dlomanov/gophkeeper/internal/infra/grpcserver"
	"github.com/dlomanov/gophkeeper/internal/infra/logging"
	"github.com/dlomanov/gophkeeper/internal/infra/migrator"
	"go.uber.org/zap"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run(ctx context.Context, config *config.Config) error {
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
	if c, err = deps.NewContainer(logger, config); err != nil {
		logger.Error("failed to init container", zap.Error(err))
		return err
	}
	defer closeContainer(c)
	if err = upMigrations(c); err != nil {
		return err
	}

	grpcsrv := startGRPC(c)
	wait(ctx, c, grpcsrv)
	shutdownGRPC(c, grpcsrv)

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

func startGRPC(c *deps.Container) *grpcserver.Server {
	s := grpcserver.New(
		grpcserver.Addr(c.Config.Address),
		grpcserver.ShutdownTimeout(15*time.Second),
		grpc.GetOptions(c),
	)
	grpc.UseServices(s, c)
	c.Logger.Debug("GRPC-server started")
	return s
}

func wait(
	ctx context.Context,
	c *deps.Container,
	grpcserv *grpcserver.Server,
) {
	terminate := make(chan os.Signal, 1)
	signal.Notify(terminate, syscall.SIGINT, syscall.SIGTERM)
	select {
	case <-ctx.Done():
		c.Logger.Info("cached cancellation", zap.Error(ctx.Err()))
	case s := <-terminate:
		c.Logger.Info("cached terminate signal", zap.String("signal", s.String()))
	case err := <-grpcserv.Notify():
		c.Logger.Error("GRPC-server notified error", zap.Error(err))
	}
}

func shutdownGRPC(c *deps.Container, s *grpcserver.Server) {
	c.Logger.Debug("GRPC-server shutdown")
	if err := s.Shutdown(); err != nil {
		c.Logger.Error("GRPC-server shutdown error", zap.Error(err))
		return
	}
	c.Logger.Debug("GRPC-server shutdown - ok")
}
