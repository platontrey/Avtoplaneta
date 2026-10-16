package grpc

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"template-service/internal/service"
)

type Server struct {
	grpcServer  *grpc.Server
	itemService service.ItemService
	port        string
}

func NewServer(itemService service.ItemService, port string) *Server {
	s := grpc.NewServer()

	// 1. Стандартный gRPC Health Check
	healthServer := health.NewServer()
	healthServer.SetServingStatus("template-service", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, healthServer)

	// 2. Reflection для утилит вроде grpcurl в среде разработки
	reflection.Register(s)

	return &Server{
		grpcServer:  s,
		itemService: itemService,
		port:        port,
	}
}

// Start запускает gRPC-сервер и блокирует выполнение до отмены контекста
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("ошибка открытия gRPC порта %s: %w", s.port, err)
	}

	errChan := make(chan error, 1)
	go func() {
		logrus.WithField("port", s.port).Info("gRPC-сервер запущен")
		if err := s.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logrus.Info("Graceful остановка gRPC-сервера...")
		s.grpcServer.GracefulStop()
		logrus.Info("gRPC-сервер остановлен")
		return nil
	case err := <-errChan:
		return fmt.Errorf("ошибка gRPC-сервера: %w", err)
	}
}
