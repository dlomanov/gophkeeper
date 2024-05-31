package interceptor

import (
	"context"
	"fmt"
	sharedmd "github.com/dlomanov/gophkeeper/internal/apps/shared/md"
	"github.com/dlomanov/gophkeeper/internal/entities"
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

func Auth(tokener Tokener) grpc.UnaryServerInterceptor {
	return auth.UnaryServerInterceptor(func(ctx context.Context) (context.Context, error) {
		t, err := auth.AuthFromMD(ctx, sharedmd.Schema)
		if err != nil {
			return ctx, fmt.Errorf("auth: failed to get token: %w", err)
		}
		token := entities.Token(t)
		userID, err := tokener.GetUserID(ctx, token)
		if err != nil {
			return ctx, fmt.Errorf("auth: failed to get user ID from token: %w", err)
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
