package grpcclient

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

// DefaultDialOptions возвращает промышленный набор опций для gRPC-клиентов Avtoplaneta:
// 1. Insecure credentials (для защищенной внутренней сети Traefik/Docker).
// 2. Keepalive параметры для предотвращения зависания полузакрытых TCP-соединений.
func DefaultDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                30 * time.Second,
			Timeout:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	}
}

// NewClient создает *grpc.ClientConn с автоматическим keepalive и безопасными дефолтами
func NewClient(target string, extraOpts ...grpc.DialOption) (*grpc.ClientConn, error) {
	opts := append(DefaultDialOptions(), extraOpts...)
	conn, err := grpc.NewClient(target, opts...)
	if err != nil {
		return nil, fmt.Errorf("gRPC connect to %s error: %w", target, err)
	}
	return conn, nil
}
