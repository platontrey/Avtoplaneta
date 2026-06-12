package main

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
)

var (
	authGRPCClient   authv1.AuthServiceClient
	ordersGRPCClient ordersv1.OrdersServiceClient
	authConn         *grpc.ClientConn
	ordersConn       *grpc.ClientConn
)

func initGRPCClients() {
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

func closeGRPCClients() {
	if authConn != nil {
		authConn.Close()
	}
	if ordersConn != nil {
		ordersConn.Close()
	}
}

func logUserActivityGRPC(ctx context.Context, userIDStr, userName, userEmail, action, resourceType, details string, resourceID uint32) {
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

func getMonthlySalesGRPC(ctx context.Context) ([]MonthlySales, error) {
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

	sales := make([]MonthlySales, len(resp.Entries))
	for i, e := range resp.Entries {
		sales[i] = MonthlySales{
			Month: e.Month,
			Sales: e.Sales,
		}
	}

	return sales, nil
}

func getUserINNGRPC(ctx context.Context, userID uint32) (string, error) {
	if authGRPCClient == nil {
		return "", errGRPCNotInitialized()
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := authGRPCClient.GetUser(ctx, &authv1.GetUserRequest{Id: userID})
	if err != nil {
		return "", nil
	}

	return resp.Inn, nil
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
