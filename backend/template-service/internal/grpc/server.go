package grpc

import (
	"context"

	"avtoplaneta/pkg/grpcserver"
	"template-service/internal/service"
)

type Server struct {
	server      *grpcserver.Server
	itemService service.ItemService
}

func NewServer(itemService service.ItemService, port string) *Server {
	srv := grpcserver.New("template-service", port)

	// Здесь регистрируются сгенерированные Protobuf сервисы:
	// templatev1.RegisterTemplateServiceServer(srv.RawServer(), newHandler(itemService))

	return &Server{
		server:      srv,
		itemService: itemService,
	}
}

// Start запускает gRPC-сервер с поддержкой graceful shutdown
func (s *Server) Start(ctx context.Context) error {
	return s.server.Run(ctx)
}
