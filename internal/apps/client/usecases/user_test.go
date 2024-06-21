package usecases_test

import (
	"context"
	"github.com/dlomanov/gophkeeper/internal/apps/client/entities"
	"github.com/dlomanov/gophkeeper/internal/apps/client/infra/services/mem"
	"github.com/dlomanov/gophkeeper/internal/apps/client/usecases"
	"github.com/dlomanov/gophkeeper/internal/apps/shared/proto"
	"github.com/dlomanov/gophkeeper/internal/apps/shared/proto/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"go.uber.org/zap/zaptest"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestUserUC_SignUp(t *testing.T) {
	tests := []struct {
		name        string
		response    *proto.SignUpUserResponse
		responseErr error
		wantErr     require.ErrorAssertionFunc
		wantCache   require.ValueAssertionFunc
	}{
		{
			name:        "user exists error",
			response:    &proto.SignUpUserResponse{},
			responseErr: status.Error(codes.AlreadyExists, ""),
			wantErr:     exactErr(entities.ErrUserExists),
			wantCache:   nothing,
		},
		{
			name:        "user invalid creds error",
			response:    &proto.SignUpUserResponse{},
			responseErr: status.Error(codes.InvalidArgument, ""),
			wantErr:     exactErr(entities.ErrUserCredsInvalid),
			wantCache:   nothing,
		},
		{
			name:        "internal server error",
			response:    &proto.SignUpUserResponse{},
			responseErr: status.Error(codes.Internal, ""),
			wantErr:     exactErr(entities.ErrServerInternal),
			wantCache:   nothing,
		},
		{
			name:        "ok",
			response:    &proto.SignUpUserResponse{Token: "token"},
			responseErr: nil,
			wantErr:     require.NoError,
			wantCache: func(t require.TestingT, value any, i ...any) {
				c := value.(*mem.Cache)
				login, ok := c.GetString("login")
				require.True(t, ok, "login should be in cache")
				require.Equal(t, "login", login)
				token, ok := c.GetString("token")
				require.True(t, ok, "token should be in cache")
				require.Equal(t, "token", token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := mocks.NewMockUserServiceClient(ctrl)
			client.EXPECT().SignUp(
				gomock.Any(),
				gomock.Any()).Return(tt.response, tt.responseErr)
			cache := mem.NewCache()
			sut := usecases.NewUserUC(
				zaptest.NewLogger(t),
				cache,
				client)

			err := sut.SignUp(context.Background(), entities.SignUpUserRequest{
				Login:    "login",
				Password: "password",
			})
			tt.wantErr(t, err)
		})
	}
}

func TestUserUC_SignIn(t *testing.T) {
	tests := []struct {
		name        string
		response    *proto.SignInUserResponse
		responseErr error
		wantErr     require.ErrorAssertionFunc
		wantCache   require.ValueAssertionFunc
	}{
		{
			name:        "invalid creds error",
			response:    &proto.SignInUserResponse{},
			responseErr: status.Error(codes.InvalidArgument, ""),
			wantErr:     exactErr(entities.ErrUserCredsInvalid),
			wantCache:   nothing,
		},
		{
			name:        "invalid creds error",
			response:    &proto.SignInUserResponse{},
			responseErr: status.Error(codes.Unauthenticated, ""),
			wantErr:     exactErr(entities.ErrUserCredsInvalid),
			wantCache:   nothing,
		},
		{
			name:        "internal server error",
			response:    &proto.SignInUserResponse{},
			responseErr: status.Error(codes.Internal, ""),
			wantErr:     exactErr(entities.ErrServerInternal),
			wantCache:   nothing,
		},
		{
			name:        "ok",
			response:    &proto.SignInUserResponse{Token: "token"},
			responseErr: nil,
			wantErr:     require.NoError,
			wantCache: func(t require.TestingT, value any, i ...any) {
				c := value.(*mem.Cache)
				login, ok := c.GetString("login")
				require.True(t, ok, "login should be in cache")
				require.Equal(t, "login", login)
				token, ok := c.GetString("token")
				require.True(t, ok, "token should be in cache")
				require.Equal(t, "token", token)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()
			client := mocks.NewMockUserServiceClient(ctrl)
			client.EXPECT().SignIn(
				gomock.Any(),
				gomock.Any()).Return(tt.response, tt.responseErr)
			cache := mem.NewCache()
			sut := usecases.NewUserUC(
				zaptest.NewLogger(t),
				cache,
				client)

			err := sut.SignIn(context.Background(), entities.SignInUserRequest{
				Login:    "login",
				Password: "password",
			})
			tt.wantErr(t, err)
		})
	}
}

func exactErr(err error) require.ErrorAssertionFunc {
	return func(t require.TestingT, err2 error, i ...any) {
		require.ErrorIs(t, err2, err, i)
	}
}

func nothing(_ require.TestingT, _ any, _ ...any) {}
