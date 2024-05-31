package services

import (
	"context"
	"errors"
	"github.com/dlomanov/gophkeeper/internal/apps/server/usecases"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/core/apperrors"
	"github.com/dlomanov/gophkeeper/internal/entities"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	_ pb.UserServiceServer         = (*UserService)(nil)
	_ auth.ServiceAuthFuncOverride = (*UserService)(nil)
)

type UserService struct {
	pb.UnimplementedUserServiceServer
	logger *zap.Logger
	userUC *usecases.UserUC
}

func NewUserService(
	logger *zap.Logger,
	userUC *usecases.UserUC,
) *UserService {
	return &UserService{
		logger: logger,
		userUC: userUC,
	}
}

func (s *UserService) SignUp(ctx context.Context, request *pb.SignUpUserRequest) (*pb.SignUpUserResponse, error) {
	creds := entities.Creds{
		Login: entities.Login(request.Login),
		Pass:  entities.Pass(request.Password),
	}

	token, err := s.userUC.SignUp(ctx, creds)
	if err != nil {
		s.logger.Debug("failed to sign up", zap.Error(err))
	}
	var invalid *apperrors.AppErrorInvalid
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.SignUpUserResponse{Token: string(token)}, nil
}

func (s *UserService) SignIn(ctx context.Context, request *pb.SignInUserRequest) (*pb.SignInUserResponse, error) {
	creds := entities.Creds{
		Login: entities.Login(request.Login),
		Pass:  entities.Pass(request.Password),
	}

	token, err := s.userUC.SignIn(ctx, creds)
	if err != nil {
		s.logger.Debug("failed to sign in", zap.Error(err))
	}
	var (
		invalid  *apperrors.AppErrorInvalid
		notFound *apperrors.AppErrorNotFound
	)
	switch {
	case errors.As(err, &invalid):
		return nil, status.Error(codes.InvalidArgument, err.Error())
	case errors.As(err, &notFound):
		return nil, status.Error(codes.Unauthenticated, err.Error())
	case err != nil:
		return nil, status.Error(codes.Internal, "internal server error")
	}

	return &pb.SignInUserResponse{Token: string(token)}, nil
}

func (s *UserService) AuthFuncOverride(ctx context.Context, _ string) (context.Context, error) {
	return ctx, nil
}
