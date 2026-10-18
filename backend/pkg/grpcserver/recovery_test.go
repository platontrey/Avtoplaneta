package grpcserver

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestRecoveryUnaryInterceptor_Success(t *testing.T) {
	interceptor := RecoveryUnaryInterceptor()

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/TestMethod"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return "success", nil
	}

	resp, err := interceptor(context.Background(), "request", info, handler)
	require.NoError(t, err)
	assert.Equal(t, "success", resp)
}

func TestRecoveryUnaryInterceptor_CatchesPanic(t *testing.T) {
	interceptor := RecoveryUnaryInterceptor()

	info := &grpc.UnaryServerInfo{FullMethod: "/test.Service/PanicMethod"}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		panic("unexpected nil pointer dereference")
	}

	resp, err := interceptor(context.Background(), "request", info, handler)
	assert.Nil(t, resp)
	require.Error(t, err)

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.Internal, st.Code())
	assert.Contains(t, st.Message(), "внутренняя ошибка сервера")
}
