package grpcclient

import (
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

const defaultRetryServiceConfig = `{
	"methodConfig": [{
		"name": [{"service": ""}],
		"retryPolicy": {
			"maxAttempts": 5,
			"initialBackoff": "0.1s",
			"maxBackoff": "1s",
			"backoffMultiplier": 2.0,
			"retryableStatusCodes": ["UNAVAILABLE", "RESOURCE_EXHAUSTED"]
		}
	}]
}`

// DefaultDialOptions возвращает промышленный набор опций для gRPC-клиентов Avtoplaneta:
// 1. Insecure credentials (для защищенной внутренней сети Traefik/Docker).
// 2. Встроенный Retry Policy с exponential backoff для прозрачного переживания рестартов сервисов.
// 3. Keepalive параметры для предотвращения зависания полузакрытых TCP-соединений.
func DefaultDialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(defaultRetryServiceConfig),
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
