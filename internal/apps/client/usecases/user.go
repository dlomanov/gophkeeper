package usecases

import (
	"context"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	pb "github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	UserUC struct {
		logger     *zap.Logger
		cache      *mem.Cache
		userClient pb.UserServiceClient
	}
	SignUpUserRequest struct {
		Login    string
		Password string
	}
	SignInUserRequest struct {
		Login    string
		Password string
	}
)

func NewUserUC(
	logger *zap.Logger,
	cache *mem.Cache,
	userClient pb.UserServiceClient,
) *UserUC {
	return &UserUC{
		logger:     logger,
		cache:      cache,
		userClient: userClient,
	}
}

func (uc *UserUC) SignUp(
	ctx context.Context,
	request SignUpUserRequest,
) error {
	resp, err := uc.userClient.SignUp(ctx, &pb.SignUpUserRequest{
		Login:    request.Login,
		Password: request.Password,
	})
	switch {
	case status.Code(err) == codes.AlreadyExists:
		return entities.ErrUserExists
	case status.Code(err) == codes.InvalidArgument:
		return entities.ErrUserCredsInvalid
	case err != nil:
		return fmt.Errorf("user_sign_up: internal server error: %w", err)
	}
	uc.cache.SetString("login", request.Login)
	uc.cache.SetString("token", resp.Token)
	return nil
}

func (uc *UserUC) SignIn(
	ctx context.Context,
	request SignInUserRequest,
) (err error) {
	resp, err := uc.userClient.SignIn(ctx, &pb.SignInUserRequest{
		Login:    request.Login,
		Password: request.Password,
	})
	switch {
	case status.Code(err) == codes.InvalidArgument:
		return fmt.Errorf("user_sign_up: invalid argument: %w", err)
	case status.Code(err) == codes.Unauthenticated:
		return fmt.Errorf("user_sign_up: unauthenticated: %w", err)
	case err != nil:
		return fmt.Errorf("user_sign_up: internal server error: %w", err)
	}
	uc.cache.SetString("login", request.Login)
	uc.cache.SetString("token", resp.Token)
	return nil
}
