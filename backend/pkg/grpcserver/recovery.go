package grpcserver

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
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
				logrus.WithFields(logrus.Fields{
					"method": info.FullMethod,
					"panic":  fmt.Sprintf("%v", r),
					"stack":  string(debug.Stack()),
				}).Error("gRPC handler panic recovered")

				err = status.Errorf(codes.Internal, "внутренняя ошибка сервера")
			}
		}()

		return handler(ctx, req)
	}
}
