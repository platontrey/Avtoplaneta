package main

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	authv1 "avtoplaneta/gen/auth/v1"
)

type authGRPCServer struct {
	authv1.UnimplementedAuthServiceServer
	authService AuthService
	config      *Config
}

func NewAuthGRPCServer(authService AuthService, config *Config) *authGRPCServer {
	return &authGRPCServer{
		authService: authService,
		config:      config,
	}
}

func (s *authGRPCServer) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	if req.Authorization != "" {
		tokenString := req.Authorization
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}
		claims, err := ValidateJWTToken(tokenString, s.config.JWTSecret)
		if err != nil {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
		user, err := s.authService.GetCurrentUser(claims.UserID)
		if err != nil {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
		return &authv1.ValidateSessionResponse{
			Valid: true,
			User:  userToProto(user),
		}, nil
	}

	if req.SessionCookie != "" {
		userID, err := validateSessionCookie(req.SessionCookie)
		if err != nil {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
		user, err := s.authService.GetCurrentUser(userID)
		if err != nil {
			return &authv1.ValidateSessionResponse{Valid: false}, nil
		}
		return &authv1.ValidateSessionResponse{
			Valid: true,
			User:  userToProto(user),
		}, nil
	}

	return &authv1.ValidateSessionResponse{Valid: false}, nil
}

func (s *authGRPCServer) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.User, error) {
	user, err := s.authService.GetCurrentUser(int64(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "пользователь не найден: %v", err)
	}
	return userToProto(user), nil
}

func (s *authGRPCServer) GetUsers(ctx context.Context, req *authv1.GetUsersRequest) (*authv1.UserList, error) {
	users, err := s.authService.GetUsers()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить пользователей: %v", err)
	}

	protoUsers := make([]*authv1.User, len(users))
	for i, u := range users {
		protoUsers[i] = &authv1.User{
			Id:       uint32(u.ID),
			Email:    u.Email,
			Name:     u.Name,
			Role:     u.Role,
			Initials: u.Initials,
			Inn:      u.INN,
			Provider: u.Provider,
		}
	}

	return &authv1.UserList{Users: protoUsers}, nil
}

func (s *authGRPCServer) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.User, error) {
	createReq := CreateUserRequest{
		Email:    req.Email,
		Name:     req.Name,
		Password: req.Password,
		Role:     req.Role,
		Initials: req.Initials,
		INN:      req.Inn,
	}

	user, err := s.authService.CreateUser(createReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return userToProto(user), nil
}

func (s *authGRPCServer) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.User, error) {
	updateReq := UpdateUserRequest{
		Name:     req.Name,
		Email:    req.Email,
		Role:     req.Role,
		Initials: req.Initials,
		INN:      req.Inn,
	}

	user, err := s.authService.UpdateUser(int64(req.Id), updateReq)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return userToProto(user), nil
}

func (s *authGRPCServer) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if err := s.authService.DeleteUser(int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return &authv1.DeleteUserResponse{}, nil
}

func (s *authGRPCServer) LogActivity(ctx context.Context, req *authv1.LogActivityRequest) (*authv1.LogActivityResponse, error) {
	user, err := s.authService.GetCurrentUser(int64(req.UserId))
	if err != nil {
		user = &User{
			ID:    int64(req.UserId),
			Name:  req.UserName,
			Email: req.UserEmail,
		}
	}

	var resourceID *int64
	if req.ResourceId != 0 {
		rid := int64(req.ResourceId)
		resourceID = &rid
	}

	if err := s.authService.LogUserActivity(user, req.Action, req.ResourceType, resourceID, req.Details, req.IpAddress, req.UserAgent); err != nil {
		logrus.WithError(err).Warn("Failed to log activity via gRPC")
	}

	return &authv1.LogActivityResponse{}, nil
}

func (s *authGRPCServer) GetActivityLogs(ctx context.Context, req *authv1.GetActivityLogsRequest) (*authv1.ActivityLogList, error) {
	filters := ActivityLogFilters{
		Action:       req.Action,
		ResourceType: req.ResourceType,
		Limit:        int(req.Limit),
		Offset:       int(req.Offset),
	}

	if req.UserId != 0 {
		uid := int64(req.UserId)
		filters.UserID = &uid
	}
	if req.StartDate != nil {
		t := req.StartDate.AsTime()
		filters.StartDate = &t
	}
	if req.EndDate != nil {
		t := req.EndDate.AsTime()
		filters.EndDate = &t
	}

	if filters.Limit == 0 {
		filters.Limit = 100
	}

	logs, err := s.authService.GetUserActivityLogs(filters)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить логи: %v", err)
	}

	protoLogs := make([]*authv1.ActivityLog, len(logs))
	for i, l := range logs {
		protoLog := &authv1.ActivityLog{
			Id:           uint32(l.ID),
			UserId:       uint32(l.UserID),
			UserName:     l.UserName,
			UserEmail:    l.UserEmail,
			Action:       l.Action,
			ResourceType: l.ResourceType,
			Details:      l.Details,
			IpAddress:    l.IPAddress,
			UserAgent:    l.UserAgent,
			CreatedAt:    timestamppb.New(l.CreatedAt),
		}
		if l.ResourceID != nil {
			protoLog.ResourceId = uint32(*l.ResourceID)
		}
		protoLogs[i] = protoLog
	}

	return &authv1.ActivityLogList{
		Logs:  protoLogs,
		Total: int32(len(protoLogs)),
	}, nil
}

func (s *authGRPCServer) GetServerStatus(ctx context.Context, req *authv1.GetServerStatusRequest) (*authv1.ServerStatus, error) {
	totalUsers, err := s.authService.GetTotalUsersCount()
	if err != nil {
		totalUsers = 0
	}

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return &authv1.ServerStatus{
		Status:        "ok",
		UptimeSeconds: 0,
		TotalUsers:    int32(totalUsers),
		MemoryUsageMb: float64(memStats.Alloc) / 1024 / 1024,
		Goroutines:    int32(runtime.NumGoroutine()),
	}, nil
}

func (s *authGRPCServer) GetServerLogs(ctx context.Context, req *authv1.GetServerLogsRequest) (*authv1.ServerLogsResponse, error) {
	logs := []*authv1.ServerLog{
		{
			Timestamp: time.Now().Format(time.RFC3339),
			Level:     "INFO",
			Message:   "Server started successfully",
			Source:    "auth-service",
		},
		{
			Timestamp: time.Now().Add(-time.Minute).Format(time.RFC3339),
			Level:     "INFO",
			Message:   "All services are running",
			Source:    "auth-service",
		},
	}
	return &authv1.ServerLogsResponse{Logs: logs}, nil
}

func userToProto(u *User) *authv1.User {
	return &authv1.User{
		Id:       uint32(u.ID),
		Email:    u.Email,
		Name:     u.Name,
		Role:     u.Role,
		Initials: u.Initials,
		Inn:      u.INN,
		Provider: u.Provider,
	}
}

func validateSessionCookie(cookieHeader string) (int64, error) {
	fakeReq, _ := http.NewRequest("GET", "/", nil)
	fakeReq.Header.Set("Cookie", cookieHeader)

	session, err := store.Get(fakeReq, "auth-session")
	if err != nil {
		return 0, fmt.Errorf("invalid session: %w", err)
	}

	val, ok := session.Values["user_id"]
	if !ok || val == nil {
		return 0, fmt.Errorf("no user_id in session")
	}

	userID, ok := getUserIDFromSessionValue(val)
	if !ok {
		return 0, fmt.Errorf("invalid user_id type in session")
	}

	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 {
			return 0, fmt.Errorf("session expired")
		}
	}

	return userID, nil
}

func StartGRPCServer(authService AuthService, config *Config, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
		),
	)

	authv1.RegisterAuthServiceServer(srv, NewAuthGRPCServer(authService, config))

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("auth.v1.AuthService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)

	logrus.WithField("port", port).Info("gRPC server listening")
	return srv.Serve(lis)
}

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
