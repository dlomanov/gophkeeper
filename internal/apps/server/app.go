package server

import (
	"context"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/server/config"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/deps"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/grpcserver"
	"github.com/dlomanov/gophkeeper/internal/infra/logging"
	"go.uber.org/zap"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
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
		Level: config.LogLevel,
		Type:  config.LogType,
	}); err != nil {
		return err
	}
	defer func(logger *zap.Logger) { _ = logger.Sync() }(logger)
	if c, err = deps.NewContainer(logger, config); err != nil {
		logger.Error("failed to init container", zap.Error(err))
		return err
	}
	defer closeContainer(c)

	grpcsrv := startGRPC(ctx, c)
	wait(ctx, c, grpcsrv)
	shutdownGRPC(c, grpcsrv)

	return nil
}

func closeContainer(c *deps.Container) {
	if err := c.Close(); err != nil {
		c.Logger.Error("failed to close container", zap.Error(err))
	}
}

func startGRPC(ctx context.Context, c *deps.Container) *grpcserver.Server {
	opts := []grpcserver.Option{
		grpcserver.Addr(c.Config.Address),
		grpcserver.ShutdownTimeout(15 * time.Second),
		grpcserver.TLSCert(c.Config.Cert, c.Config.CertKey),
		grpc.GetOptions(c),
	}

	if value := ctx.Value(grpcserver.ListenerKey); value != nil {
		c.Logger.Debug("GRPC-server custom listener detected")
		if l, ok := value.(net.Listener); ok {
			opts = append(opts, grpcserver.Listener(l))
			c.Logger.Debug("GRPC-server starts with custom listener")
		} else {
			c.Logger.Debug("GRPC-server custom listener is not net.Listener")
		}
	}

	s := grpcserver.New(opts...)
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
		c.Logger.Info("cached cancellation -> shutdown")
	case s := <-terminate:
		c.Logger.Info("cached terminate signal -> shutdown", zap.String("signal", s.String()))
	case err := <-grpcserv.Notify():
		c.Logger.Error("GRPC-server notified error -> shutdown", zap.Error(err))
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
