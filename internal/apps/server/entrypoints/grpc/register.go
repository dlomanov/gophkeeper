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
	pb.RegisterEntryServiceServer(s.Server, services.NewEntryService())
}

func GetOptions(c *deps.Container) grpcserver.Option {
	return grpcserver.ServerOptions(grpc.ChainUnaryInterceptor(
		interceptor.Auth(c.UserUC),
		interceptor.Logger(c.Logger),
		interceptor.Recovery(c.Logger),
	))
}
