package interceptor

import (
	"context"
	"errors"
	"fmt"
	"github.com/dlomanov/gophkeeper/internal/apps/server/entities"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	"github.com/google/uuid"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const UserIDKey ContextKey = "server_user_id"

type ContextKey string

func Logger(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(interceptorLogger(logger))
}

func interceptorLogger(logger *zap.Logger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		var values []zap.Field
		if len(fields)%2 == 0 {
			for i := 0; i < len(fields); i += 2 {
				if k, ok := fields[i].(string); ok {
					values = append(values, zap.Any(k, fields[i+1]))
				}
			}
		}
		switch lvl {
		case logging.LevelDebug:
			logger.Debug(msg, values...)
		case logging.LevelInfo:
			logger.Info(msg, values...)
		case logging.LevelWarn:
			logger.Warn(msg, values...)
		case logging.LevelError:
			logger.Error(msg, values...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}

func Recovery(logger *zap.Logger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p any) (err error) {
		logger.Error("cached panic", zap.Any("panic", p))
		return status.Error(codes.Internal, "internal server error")
	}))
}

type Tokener interface {
	GetUserID(ctx context.Context, token entities.Token) (uuid.UUID, error)
}

func Auth(logger *zap.Logger, tokener Tokener) grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		t, err := auth.AuthFromMD(ctx, sharedmd.Schema)
		if err != nil {
			logger.Debug("failed to get token from metadata", zap.Error(err))
			return ctx, status.Error(codes.Unauthenticated, err.Error())
		}
		token := entities.Token(t)
		userID, err := tokener.GetUserID(ctx, token)
		switch {
		case errors.Is(err, entities.ErrUserTokenInvalid):
			logger.Debug("invalid token", zap.Error(err))
			return ctx, status.Error(codes.Unauthenticated, err.Error())
		case errors.Is(err, entities.ErrUserTokenExpired):
			logger.Debug("token expired", zap.Error(err))
			return ctx, status.Error(codes.Unauthenticated, err.Error())
		case err != nil:
			logger.Error("failed to get user ID from token", zap.Error(err))
			return ctx, status.Error(codes.Internal, "internal server error")
		}
		return context.WithValue(ctx, UserIDKey, userID), nil
	})
}

func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	if v, ok := ctx.Value(UserIDKey).(uuid.UUID); ok {
		return v, true
	}
	return uuid.Nil, false
}
