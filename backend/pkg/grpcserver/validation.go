package grpcserver

import (
	"context"

	"buf.build/go/protovalidate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

var defaultValidator protovalidate.Validator

func init() {
	var err error
	defaultValidator, err = protovalidate.New()
	if err != nil {
		panic("failed to initialize protovalidate: " + err.Error())
	}
}

// ValidationUnaryInterceptor выполняет автоматическую валидацию входящих protobuf-сообщений через protovalidate
func ValidationUnaryInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if msg, ok := req.(proto.Message); ok {
			if err := defaultValidator.Validate(msg); err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "ошибка валидации запроса: %v", err)
			}
		}
		return handler(ctx, req)
	}
}
