package deps

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server"
	"go.uber.org/zap"
	"io"
)

var _ io.Closer = (*Container)(nil)

type Container struct {
	Logger *zap.Logger
	Config *server.Config
}

func NewContainer(
	logger *zap.Logger,
	config *server.Config,
) *Container {
	return &Container{
		Logger: logger,
		Config: config,
	}
}

func (c Container) Close() error {
	return nil
}
