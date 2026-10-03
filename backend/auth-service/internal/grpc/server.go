package grpc

import (
	"context"
	"fmt"
	"net"
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

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

// AuthGRPCServer реализует интерфейс authv1.AuthServiceServer
type AuthGRPCServer struct {
	authv1.UnimplementedAuthServiceServer
	authService  service.AuthService
	sessionStore security.SessionStore
	config       *config.Config
}

// NewAuthGRPCServer создает новый экземпляр AuthGRPCServer
func NewAuthGRPCServer(
	authService service.AuthService,
	sessionStore security.SessionStore,
	config *config.Config,
) *AuthGRPCServer {
	return &AuthGRPCServer{
		authService:  authService,
		sessionStore: sessionStore,
		config:       config,
	}
}

func (s *AuthGRPCServer) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	if req.Authorization != "" {
		tokenString := req.Authorization
		if strings.HasPrefix(tokenString, "Bearer ") {
			tokenString = strings.TrimPrefix(tokenString, "Bearer ")
		}
		claims, err := security.ValidateJWTToken(tokenString, s.config.JWTSecret)
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

	if req.SessionCookie != "" && s.sessionStore != nil {
		userID, err := security.ValidateSessionCookie(s.sessionStore, req.SessionCookie)
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

func (s *AuthGRPCServer) GetUser(ctx context.Context, req *authv1.GetUserRequest) (*authv1.User, error) {
	user, err := s.authService.GetCurrentUser(int64(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "пользователь не найден: %v", err)
	}
	return userToProto(user), nil
}

func (s *AuthGRPCServer) GetUsers(ctx context.Context, req *authv1.GetUsersRequest) (*authv1.UserList, error) {
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

func (s *AuthGRPCServer) CreateUser(ctx context.Context, req *authv1.CreateUserRequest) (*authv1.User, error) {
	createReq := domain.CreateUserRequest{
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

func (s *AuthGRPCServer) UpdateUser(ctx context.Context, req *authv1.UpdateUserRequest) (*authv1.User, error) {
	updateReq := domain.UpdateUserRequest{
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

func (s *AuthGRPCServer) DeleteUser(ctx context.Context, req *authv1.DeleteUserRequest) (*authv1.DeleteUserResponse, error) {
	if err := s.authService.DeleteUser(int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "%v", err)
	}
	return &authv1.DeleteUserResponse{}, nil
}

func (s *AuthGRPCServer) LogActivity(ctx context.Context, req *authv1.LogActivityRequest) (*authv1.LogActivityResponse, error) {
	user, err := s.authService.GetCurrentUser(int64(req.UserId))
	if err != nil {
		user = &domain.User{
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

func (s *AuthGRPCServer) GetActivityLogs(ctx context.Context, req *authv1.GetActivityLogsRequest) (*authv1.ActivityLogList, error) {
	filters := domain.ActivityLogFilters{
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

func (s *AuthGRPCServer) GetServerStatus(ctx context.Context, req *authv1.GetServerStatusRequest) (*authv1.ServerStatus, error) {
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

func (s *AuthGRPCServer) GetServerLogs(ctx context.Context, req *authv1.GetServerLogsRequest) (*authv1.ServerLogsResponse, error) {
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

func userToProto(u *domain.User) *authv1.User {
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

// NewGRPCServer создает настроенный gRPC сервер
func NewGRPCServer(
	authService service.AuthService,
	sessionStore security.SessionStore,
	cfg *config.Config,
) *grpc.Server {
	srv := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
		),
	)

	authv1.RegisterAuthServiceServer(srv, NewAuthGRPCServer(authService, sessionStore, cfg))

	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("auth.v1.AuthService", healthpb.HealthCheckResponse_SERVING)

	reflection.Register(srv)
	return srv
}

// StartGRPCServer запускает gRPC сервер на указанном порту
func StartGRPCServer(
	authService service.AuthService,
	sessionStore security.SessionStore,
	cfg *config.Config,
	port string,
) (*grpc.Server, error) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return nil, fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := NewGRPCServer(authService, sessionStore, cfg)

	logrus.WithField("port", port).Info("gRPC server listening")
	go func() {
		if err := srv.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			logrus.WithError(err).Error("gRPC server error")
		}
	}()

	return srv, nil
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
