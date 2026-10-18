package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	ordersv1 "avtoplaneta/gen/orders/v1"
	"avtoplaneta/pkg/grpcserver"
	"orders-service/internal/domain"
	"orders-service/internal/service"
)

// ordersGRPCServer реализует gRPC-сервер для OrdersService
type ordersGRPCServer struct {
	ordersv1.UnimplementedOrdersServiceServer
	service   service.OrdersService
	publisher domain.EventPublisher
}

// NewOrdersGRPCServer создаёт новый gRPC-сервер
func NewOrdersGRPCServer(service service.OrdersService, publisher domain.EventPublisher) *ordersGRPCServer {
	return &ordersGRPCServer{
		service:   service,
		publisher: publisher,
	}
}

// CreateOrder создаёт новый заказ
func (s *ordersGRPCServer) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.CreateOrderResponse, error) {
	createReq := domain.CreateOrderRequest{
		CustomerID:  req.GetCustomerId(),
		Part:        req.GetPart(),
		PartID:      req.GetPartId(),
		BuyerNumber: req.GetBuyerNumber(),
	}

	for _, item := range req.GetItems() {
		createReq.Items = append(createReq.Items, domain.CreateOrderItemInput{
			PartID:   item.GetPartId(),
			Quantity: int(item.GetQuantity()),
		})
	}

	order, err := s.service.CreateOrder(ctx, createReq, req.GetSellerId(), "")
	if err != nil {
		if domain.IsValidationError(err) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось создать заказ: %v", err)
	}

	return &ordersv1.CreateOrderResponse{Order: order.ToProto()}, nil
}

// GetOrders возвращает список заказов
func (s *ordersGRPCServer) GetOrders(ctx context.Context, req *ordersv1.GetOrdersRequest) (*ordersv1.GetOrdersResponse, error) {
	orders, err := s.service.GetOrders(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить заказы: %v", err)
	}

	protoOrders := make([]*ordersv1.Order, len(orders))
	for i := range orders {
		protoOrders[i] = orders[i].ToProto()
	}

	return &ordersv1.GetOrdersResponse{
		Orders: protoOrders,
		Total:  int32(len(orders)),
	}, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *ordersGRPCServer) UpdateOrderStatus(ctx context.Context, req *ordersv1.UpdateOrderStatusRequest) (*ordersv1.UpdateOrderStatusResponse, error) {
	if err := s.service.UpdateOrderStatus(ctx, req.GetId(), req.GetStatus()); err != nil {
		if domain.IsValidationError(err) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось обновить статус заказа: %v", err)
	}

	return &ordersv1.UpdateOrderStatusResponse{
		Order: &ordersv1.Order{Id: req.GetId(), Status: req.GetStatus()},
	}, nil
}

// CompleteOrder завершает заказ
func (s *ordersGRPCServer) CompleteOrder(ctx context.Context, req *ordersv1.CompleteOrderRequest) (*ordersv1.CompleteOrderResponse, error) {
	if err := s.service.CompleteOrder(ctx, req.GetId()); err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось завершить заказ: %v", err)
	}

	return &ordersv1.CompleteOrderResponse{
		Order: &ordersv1.Order{Id: req.GetId(), Status: "green"},
	}, nil
}

// DeleteOrder удаляет заказ
func (s *ordersGRPCServer) DeleteOrder(ctx context.Context, req *ordersv1.DeleteOrderRequest) (*ordersv1.DeleteOrderResponse, error) {
	if err := s.service.DeleteOrder(ctx, req.GetId()); err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось удалить заказ: %v", err)
	}

	return &ordersv1.DeleteOrderResponse{}, nil
}

// AddOrderItem добавляет позицию в заказ
func (s *ordersGRPCServer) AddOrderItem(ctx context.Context, req *ordersv1.AddOrderItemRequest) (*ordersv1.AddOrderItemResponse, error) {
	addReq := domain.AddOrderItemRequest{
		PartID:   req.GetPartId(),
		Quantity: int(req.GetQuantity()),
	}

	if err := s.service.AddOrderItem(ctx, req.GetOrderId(), addReq); err != nil {
		if domain.IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось добавить позицию: %v", err)
	}

	return &ordersv1.AddOrderItemResponse{
		Order: &ordersv1.Order{Id: req.GetOrderId()},
	}, nil
}

// GetMonthlySales возвращает продажи по месяцам
func (s *ordersGRPCServer) GetMonthlySales(ctx context.Context, req *ordersv1.GetMonthlySalesRequest) (*ordersv1.GetMonthlySalesResponse, error) {
	sales, err := s.service.GetMonthlySales(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить продажи: %v", err)
	}

	entries := make([]*ordersv1.MonthlySalesEntry, len(sales))
	for i, ms := range sales {
		entries[i] = &ordersv1.MonthlySalesEntry{
			Month: ms.Month,
			Sales: ms.Sales,
		}
	}

	return &ordersv1.GetMonthlySalesResponse{Entries: entries}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func orderToProto(o *domain.Order) *ordersv1.Order {
	return o.ToProto()
}

// ─── gRPC Server Startup ────────────────────────────────────────────────────

// StartGRPCServer запускает gRPC-сервер на указанном порту
func StartGRPCServer(service service.OrdersService, publisher domain.EventPublisher, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
			grpcserver.ValidationUnaryInterceptor(),
		),
	)

	ordersv1.RegisterOrdersServiceServer(srv, NewOrdersGRPCServer(service, publisher))

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("orders.v1.OrdersService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	logrus.WithField("port", port).Info("orders-service gRPC server listening")
	return srv.Serve(lis)
}

// ─── Interceptors ───────────────────────────────────────────────────────────

func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	fields := logrus.Fields{
		"method":   info.FullMethod,
		"duration": duration.String(),
	}
	if err != nil {
		fields["error"] = err.Error()
		logrus.WithFields(fields).Warn("gRPC call failed")
	} else if duration > 100*time.Millisecond {
		logrus.WithFields(fields).Info("gRPC call slow")
	}

	return resp, err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).WithField("method", info.FullMethod).Error("gRPC panic recovered")
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}
