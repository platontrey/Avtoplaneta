package httpserver

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestServerGracefulShutdown(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("pong"))
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	serverErr := make(chan error, 1)
	go func() {
		serverErr <- Run(ctx, "0", mux, WithShutdownTimeout(2*time.Second))
	}()

	// Даём серверу запуститься и отменяем контекст
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-serverErr:
		require.NoError(t, err)
	case <-time.After(3 * time.Second):
		t.Fatal("Server shutdown timed out")
	}
}
