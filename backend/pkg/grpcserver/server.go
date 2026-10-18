package grpcserver

import (
	"context"
	"fmt"
	"net"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server обёртка над стандартным grpc.Server с health check, reflection и graceful shutdown
type Server struct {
	grpcServer   *grpc.Server
	healthServer *health.Server
	port         string
	serviceName  string
}

// New создает настроенный gRPC-сервер по стандартам Avtoplaneta
func New(serviceName, port string, opts ...grpc.ServerOption) *Server {
	defaultOpts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(ValidationUnaryInterceptor()),
	}
	allOpts := append(defaultOpts, opts...)
	s := grpc.NewServer(allOpts...)

	// 1. Стандартный Health Check для k8s и балансировщиков
	hs := health.NewServer()
	hs.SetServingStatus(serviceName, healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(s, hs)

	// 2. Reflection для grpcurl и отладки
	reflection.Register(s)

	return &Server{
		grpcServer:   s,
		healthServer: hs,
		port:         port,
		serviceName:  serviceName,
	}
}

// RawServer возвращает низкоуровневый *grpc.Server для регистрации protobuf-хэндлеров
func (s *Server) RawServer() *grpc.Server {
	return s.grpcServer
}

// SetServingStatus позволяет динамически менять статус сервиса (SERVING, NOT_SERVING)
func (s *Server) SetServingStatus(status healthpb.HealthCheckResponse_ServingStatus) {
	s.healthServer.SetServingStatus(s.serviceName, status)
}

// Run запускает прослушивание порта и блокирует выполнение до отмены контекста
func (s *Server) Run(ctx context.Context) error {
	lis, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		return fmt.Errorf("gRPC server listen on :%s error: %w", s.port, err)
	}

	errChan := make(chan error, 1)
	go func() {
		logrus.WithFields(logrus.Fields{
			"port":    s.port,
			"service": s.serviceName,
		}).Info("gRPC-сервер запущен")

		if err := s.grpcServer.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logrus.WithField("service", s.serviceName).Info("Graceful остановка gRPC-сервера...")
		s.grpcServer.GracefulStop()
		logrus.WithField("service", s.serviceName).Info("gRPC-сервер успешно остановлен")
		return nil

	case err := <-errChan:
		return fmt.Errorf("gRPC-сервер завершился с ошибкой: %w", err)
	}
}
