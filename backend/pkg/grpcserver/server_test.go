package grpcserver

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
)

func TestGRPCServer_GracefulShutdown(t *testing.T) {
	srv := New("test-service", "0")
	srv.SetServingStatus(healthpb.HealthCheckResponse_NOT_SERVING)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		errChan <- srv.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errChan:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("gRPC server shutdown timed out")
	}
}
