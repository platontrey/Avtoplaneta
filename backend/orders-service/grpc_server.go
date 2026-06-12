package main

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
	"google.golang.org/protobuf/types/known/timestamppb"

	ordersv1 "avtoplaneta/gen/orders/v1"
)

// ordersGRPCServer реализует gRPC-сервер для OrdersService
type ordersGRPCServer struct {
	ordersv1.UnimplementedOrdersServiceServer
	service   OrdersService
	publisher EventPublisher
}

// NewOrdersGRPCServer создаёт новый gRPC-сервер
func NewOrdersGRPCServer(service OrdersService, publisher EventPublisher) *ordersGRPCServer {
	return &ordersGRPCServer{
		service:   service,
		publisher: publisher,
	}
}

// CreateOrder создаёт новый заказ
func (s *ordersGRPCServer) CreateOrder(ctx context.Context, req *ordersv1.CreateOrderRequest) (*ordersv1.Order, error) {
	createReq := CreateOrderRequest{
		CustomerID:  int64(req.CustomerId),
		Part:        req.Part,
		PartID:      int64(req.PartId),
		BuyerNumber: req.BuyerNumber,
	}

	for _, item := range req.Items {
		createReq.Items = append(createReq.Items, struct {
			PartID   int64 `json:"part_id"`
			Quantity int   `json:"quantity"`
		}{
			PartID:   int64(item.PartId),
			Quantity: int(item.Quantity),
		})
	}

	order, err := s.service.CreateOrder(ctx, createReq, int64(req.SellerId), "")
	if err != nil {
		if IsValidationError(err) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось создать заказ: %v", err)
	}

	return orderToProto(order), nil
}

// GetOrders возвращает список заказов
func (s *ordersGRPCServer) GetOrders(ctx context.Context, req *ordersv1.GetOrdersRequest) (*ordersv1.OrderList, error) {
	orders, err := s.service.GetOrders(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить заказы: %v", err)
	}

	protoOrders := make([]*ordersv1.Order, len(orders))
	for i, o := range orders {
		protoOrders[i] = orderToProto(&o)
	}

	return &ordersv1.OrderList{
		Orders: protoOrders,
		Total:  int32(len(orders)),
	}, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *ordersGRPCServer) UpdateOrderStatus(ctx context.Context, req *ordersv1.UpdateOrderStatusRequest) (*ordersv1.Order, error) {
	if err := s.service.UpdateOrderStatus(ctx, int64(req.Id), req.Status); err != nil {
		if IsValidationError(err) {
			return nil, status.Errorf(codes.InvalidArgument, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось обновить статус заказа: %v", err)
	}

	return &ordersv1.Order{Id: req.Id, Status: req.Status}, nil
}

// CompleteOrder завершает заказ
func (s *ordersGRPCServer) CompleteOrder(ctx context.Context, req *ordersv1.CompleteOrderRequest) (*ordersv1.Order, error) {
	if err := s.service.CompleteOrder(ctx, int64(req.Id)); err != nil {
		if IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось завершить заказ: %v", err)
	}

	return &ordersv1.Order{Id: req.Id, Status: "green"}, nil
}

// DeleteOrder удаляет заказ
func (s *ordersGRPCServer) DeleteOrder(ctx context.Context, req *ordersv1.DeleteOrderRequest) (*ordersv1.DeleteOrderResponse, error) {
	if err := s.service.DeleteOrder(ctx, int64(req.Id)); err != nil {
		if IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось удалить заказ: %v", err)
	}

	return &ordersv1.DeleteOrderResponse{}, nil
}

// AddOrderItem добавляет позицию в заказ
func (s *ordersGRPCServer) AddOrderItem(ctx context.Context, req *ordersv1.AddOrderItemRequest) (*ordersv1.Order, error) {
	addReq := AddOrderItemRequest{
		PartID:   int64(req.PartId),
		Quantity: int(req.Quantity),
	}

	if err := s.service.AddOrderItem(ctx, int64(req.OrderId), addReq); err != nil {
		if IsNotFoundError(err) {
			return nil, status.Errorf(codes.NotFound, "%v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось добавить позицию: %v", err)
	}

	return &ordersv1.Order{Id: req.OrderId}, nil
}

// GetMonthlySales возвращает продажи по месяцам
func (s *ordersGRPCServer) GetMonthlySales(ctx context.Context, req *ordersv1.GetMonthlySalesRequest) (*ordersv1.MonthlySalesResponse, error) {
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

	return &ordersv1.MonthlySalesResponse{Entries: entries}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func orderToProto(o *Order) *ordersv1.Order {
	proto := &ordersv1.Order{
		Id:                 uint32(o.ID),
		CustomerId:         int32(o.CustomerID),
		SellerId:           uint32(o.SellerID),
		Seller:             o.Seller,
		Part:               o.Part,
		PartId:             uint32(o.PartID),
		Location:           o.Location,
		BuyerNumber:        o.BuyerNumber,
		Status:             o.Status,
		StatusText:         o.StatusText,
		AutoDeleted:        o.AutoDeleted,
		CreatedAt:          timestamppb.New(o.CreatedAt),
		CreatedAtFormatted: o.CreatedAtFormatted,
		TimeAgo:            o.TimeAgo,
	}

	for _, item := range o.Items {
		proto.Items = append(proto.Items, &ordersv1.OrderItem{
			Id:       uint32(item.ID),
			OrderId:  uint32(item.OrderID),
			PartId:   uint32(item.PartID),
			Quantity: int32(item.Quantity),
			Price:    item.Price,
		})
	}

	return proto
}

// ─── gRPC Server Startup ────────────────────────────────────────────────────

// StartGRPCServer запускает gRPC-сервер на указанном порту
func StartGRPCServer(service OrdersService, publisher EventPublisher, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
		),
	)

	ordersv1.RegisterOrdersServiceServer(srv, NewOrdersGRPCServer(service, publisher))

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("orders.v1.OrdersService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	logrus.WithField("port", port).Info("gRPC server listening")
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
