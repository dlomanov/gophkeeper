package grpc

import (
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc/interceptor"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entrypoints/grpc/services"
	"github.com/dlomanov/gophkeeper/internal/apps/server/infra/deps"
	grpcserver2 "github.com/dlomanov/gophkeeper/internal/apps/server/infra/grpcserver"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"google.golang.org/grpc"
)

func UseServices(s *grpcserver2.Server, c *deps.Container) {
	pb.RegisterUserServiceServer(s.Server, services.NewUserService(c.Logger, c.UserUC))
	pb.RegisterEntryServiceServer(s.Server, services.NewEntryService(c.Logger, c.EntryUC))
}

func GetOptions(c *deps.Container) grpcserver2.Option {
	return grpcserver2.ServerOptions(grpc.ChainUnaryInterceptor(
		interceptor.Auth(c.UserUC),
		interceptor.Logger(c.Logger),
		interceptor.Recovery(c.Logger),
	))
}
