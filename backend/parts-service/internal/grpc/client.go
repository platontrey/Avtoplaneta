package grpc

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authv1 "avtoplaneta/gen/auth/v1"
	ordersv1 "avtoplaneta/gen/orders/v1"
	"parts-service/internal/domain"
)

var (
	authGRPCClient   authv1.AuthServiceClient
	ordersGRPCClient ordersv1.OrdersServiceClient
	authConn         *grpc.ClientConn
	ordersConn       *grpc.ClientConn
)

// InitGRPCClients инициализирует подключения к auth-service и orders-service
func InitGRPCClients() {
	authAddr := getGRPCAddr("AUTH_GRPC_ADDR", "localhost:9083")
	ordersAddr := getGRPCAddr("ORDERS_GRPC_ADDR", "localhost:9082")

	var err error
	authConn, err = grpc.NewClient(authAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create auth gRPC connection")
	} else {
		authGRPCClient = authv1.NewAuthServiceClient(authConn)
		logrus.WithField("addr", authAddr).Info("gRPC client to auth-service initialized")
	}

	ordersConn, err = grpc.NewClient(ordersAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create orders gRPC connection")
	} else {
		ordersGRPCClient = ordersv1.NewOrdersServiceClient(ordersConn)
		logrus.WithField("addr", ordersAddr).Info("gRPC client to orders-service initialized")
	}
}

// CloseGRPCClients закрывает все открытые gRPC соединения
func CloseGRPCClients() {
	if authConn != nil {
		authConn.Close()
	}
	if ordersConn != nil {
		ordersConn.Close()
	}
}

// GetAuthClient возвращает клиент auth-service
func GetAuthClient() authv1.AuthServiceClient {
	return authGRPCClient
}

// GetOrdersClient возвращает клиент orders-service
func GetOrdersClient() ordersv1.OrdersServiceClient {
	return ordersGRPCClient
}

// LogUserActivityGRPC логирует активность пользователя в auth-service
func LogUserActivityGRPC(ctx context.Context, userIDStr, userName, userEmail, action, resourceType, details string, resourceID uint32) {
	if authGRPCClient == nil {
		logrus.Warn("auth gRPC client not initialized, skipping activity log")
		return
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &authv1.LogActivityRequest{
		Action:       action,
		ResourceType: resourceType,
		ResourceId:   resourceID,
		Details:      details,
		UserName:     userName,
		UserEmail:    userEmail,
	}

	if uid, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
		req.UserId = uint32(uid)
	}

	if _, err := authGRPCClient.LogActivity(ctx, req); err != nil {
		st := status.Convert(err)
		logrus.WithFields(logrus.Fields{
			"code":    st.Code().String(),
			"message": st.Message(),
		}).Warn("Failed to log activity via gRPC")
	}
}

// GetMonthlySalesGRPC запрашивает месячные продажи из orders-service
func GetMonthlySalesGRPC(ctx context.Context) ([]domain.MonthlySales, error) {
	if ordersGRPCClient == nil {
		return nil, errGRPCNotInitialized()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := ordersGRPCClient.GetMonthlySales(ctx, &ordersv1.GetMonthlySalesRequest{})
	if err != nil {
		st := status.Convert(err)
		if st.Code() == codes.Unavailable {
			logrus.Warn("orders-service is unavailable via gRPC")
			return nil, errGRPCNotInitialized()
		}
		return nil, nil
	}

	sales := make([]domain.MonthlySales, len(resp.Entries))
	for i, e := range resp.Entries {
		sales[i] = domain.MonthlySales{
			Month: e.Month,
			Sales: e.Sales,
		}
	}

	return sales, nil
}

func errGRPCNotInitialized() error {
	return status.Error(codes.Unavailable, "gRPC client not initialized")
}

func getGRPCAddr(envKey, defaultAddr string) string {
	if addr := os.Getenv(envKey); addr != "" {
		return addr
	}
	return defaultAddr
}
