package grpcserver

import (
	"context"
	"fmt"
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryUnaryInterceptor перехватывает паники в gRPC-обработчиках,
// логирует стек вызовов и возвращает клиенту статус codes.Internal,
// предотвращая разрыв соединений и аварийную остановку воркеров.
func RecoveryUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				slog.ErrorContext(ctx, "gRPC handler panic recovered",
					slog.String("method", info.FullMethod),
					slog.String("panic", fmt.Sprintf("%v", r)),
					slog.String("stack", string(debug.Stack())),
				)

				err = status.Errorf(codes.Internal, "внутренняя ошибка сервера")
			}
		}()

		return handler(ctx, req)
	}
}
