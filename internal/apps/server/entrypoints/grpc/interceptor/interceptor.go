package interceptor

import (
	"context"
	"fmt"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Logger(sugar *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(interceptorLogger(sugar))
}

func Recovery(sugar *zap.SugaredLogger) grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p any) (err error) {
		sugar.Error("cached panic", zap.Any("panic", p))
		return status.Error(codes.Internal, "internal server error")
	}))
}

func interceptorLogger(sugar *zap.SugaredLogger) logging.Logger {
	return logging.LoggerFunc(func(_ context.Context, lvl logging.Level, msg string, fields ...any) {
		switch lvl {
		case logging.LevelDebug:
			sugar.Debugf(msg, fields...)
		case logging.LevelInfo:
			sugar.Infof(msg, fields...)
		case logging.LevelWarn:
			sugar.Warnf(msg, fields...)
		case logging.LevelError:
			sugar.Errorf(msg, fields...)
		default:
			panic(fmt.Sprintf("unknown level %v", lvl))
		}
	})
}
