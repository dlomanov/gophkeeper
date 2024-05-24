package grpc

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc/interceptor"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc/services"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/deps"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/infra/grpcserver"
	"google.golang.org/grpc"
)

func UseServices(s *grpcserver.Server, c *deps.Container) {
	pb.RegisterUserServiceServer(s.Server, services.NewUserService(c.Logger, c.UserUC))
}

func GetOptions(c *deps.Container) grpcserver.Option {
	sugar := c.Logger.Sugar()
	return grpcserver.ServerOptions(grpc.ChainUnaryInterceptor(
		interceptor.Logger(sugar),
		interceptor.Recovery(sugar),
	))
}
