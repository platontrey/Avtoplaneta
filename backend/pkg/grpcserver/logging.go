package grpcserver

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"

	"avtoplaneta/pkg/logger"
)

// LoggingUnaryInterceptor возвращает перехватчик для структурированного логирования gRPC-запросов
func LoggingUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		duration := time.Since(start)

		if err != nil {
			slog.WarnContext(ctx, "gRPC call failed",
				slog.String("method", info.FullMethod),
				slog.Duration("duration", duration),
				logger.Err(err),
			)
		} else if duration > 100*time.Millisecond {
			slog.InfoContext(ctx, "gRPC call slow",
				slog.String("method", info.FullMethod),
				slog.Duration("duration", duration),
			)
		}

		return resp, err
	}
}
