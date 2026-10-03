package grpc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	authv1 "avtoplaneta/gen/auth/v1"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

type GRPCServerTestSuite struct {
	suite.Suite
	mockAuthService *service.MockAuthService
	grpcServer      *AuthGRPCServer
	config          *config.Config
}

func (s *GRPCServerTestSuite) SetupTest() {
	s.mockAuthService = new(service.MockAuthService)
	s.config = &config.Config{
		JWTSecret: "test-jwt-secret-with-sufficient-length-32-chars",
	}
	s.grpcServer = NewAuthGRPCServer(s.mockAuthService, nil, s.config)
}

func (s *GRPCServerTestSuite) TearDownTest() {
	s.mockAuthService.AssertExpectations(s.T())
}

func (s *GRPCServerTestSuite) TestValidateSession_ValidJWT() {
	expectedUser := &domain.User{
		ID:    10,
		Email: "admin@example.com",
		Name:  "Admin",
		Role:  "admin",
	}
	token, _, err := security.GenerateJWTTokens(expectedUser, s.config.JWTSecret)
	s.NoError(err)
	s.mockAuthService.On("GetCurrentUser", int64(10)).Return(expectedUser, nil)

	resp, err := s.grpcServer.ValidateSession(context.Background(), &authv1.ValidateSessionRequest{
		Authorization: "Bearer " + token,
	})

	s.NoError(err)
	s.True(resp.Valid)
	s.Equal(uint32(10), resp.User.Id)
	s.Equal("admin@example.com", resp.User.Email)
}

func (s *GRPCServerTestSuite) TestValidateSession_InvalidJWT() {
	resp, err := s.grpcServer.ValidateSession(context.Background(), &authv1.ValidateSessionRequest{
		Authorization: "Bearer invalid.jwt.token",
	})

	s.NoError(err)
	s.False(resp.Valid)
	s.Nil(resp.User)
}

func (s *GRPCServerTestSuite) TestValidateSession_Empty() {
	resp, err := s.grpcServer.ValidateSession(context.Background(), &authv1.ValidateSessionRequest{})

	s.NoError(err)
	s.False(resp.Valid)
	s.Nil(resp.User)
}

func (s *GRPCServerTestSuite) TestGetUser_Success() {
	expectedUser := &domain.User{
		ID:    1,
		Email: "user@example.com",
		Name:  "User",
		Role:  "operator",
	}
	s.mockAuthService.On("GetCurrentUser", int64(1)).Return(expectedUser, nil)

	resp, err := s.grpcServer.GetUser(context.Background(), &authv1.GetUserRequest{Id: 1})
	s.NoError(err)
	s.Equal(uint32(1), resp.Id)
	s.Equal("user@example.com", resp.Email)
}

func (s *GRPCServerTestSuite) TestGetUser_NotFound() {
	s.mockAuthService.On("GetCurrentUser", int64(999)).Return(nil, assert.AnError)

	_, err := s.grpcServer.GetUser(context.Background(), &authv1.GetUserRequest{Id: 999})
	s.Error(err)
}

func (s *GRPCServerTestSuite) TestGetUsers_Success() {
	users := []domain.User{
		{ID: 1, Email: "u1@example.com", Name: "U1", Role: "operator"},
		{ID: 2, Email: "u2@example.com", Name: "U2", Role: "admin"},
	}
	s.mockAuthService.On("GetUsers").Return(users, nil)

	resp, err := s.grpcServer.GetUsers(context.Background(), &authv1.GetUsersRequest{})
	s.NoError(err)
	s.Len(resp.Users, 2)
}

func (s *GRPCServerTestSuite) TestCreateUser_Success() {
	createdUser := &domain.User{
		ID:    5,
		Email: "new@example.com",
		Name:  "New",
		Role:  "operator",
	}
	s.mockAuthService.On("CreateUser", mock.MatchedBy(func(req domain.CreateUserRequest) bool {
		return req.Email == "new@example.com"
	})).Return(createdUser, nil)

	resp, err := s.grpcServer.CreateUser(context.Background(), &authv1.CreateUserRequest{
		Email: "new@example.com",
		Name:  "New",
		Role:  "operator",
	})
	s.NoError(err)
	s.Equal(uint32(5), resp.Id)
}

func (s *GRPCServerTestSuite) TestUpdateUser_Success() {
	updatedUser := &domain.User{
		ID:    5,
		Email: "updated@example.com",
		Name:  "Updated",
		Role:  "admin",
	}
	s.mockAuthService.On("UpdateUser", int64(5), mock.MatchedBy(func(req domain.UpdateUserRequest) bool {
		return req.Name == "Updated"
	})).Return(updatedUser, nil)

	resp, err := s.grpcServer.UpdateUser(context.Background(), &authv1.UpdateUserRequest{
		Id:   5,
		Name: "Updated",
	})
	s.NoError(err)
	s.Equal("Updated", resp.Name)
}

func (s *GRPCServerTestSuite) TestDeleteUser_Success() {
	s.mockAuthService.On("DeleteUser", int64(5)).Return(nil)

	_, err := s.grpcServer.DeleteUser(context.Background(), &authv1.DeleteUserRequest{Id: 5})
	s.NoError(err)
}

func (s *GRPCServerTestSuite) TestGetServerStatus() {
	s.mockAuthService.On("GetTotalUsersCount").Return(42, nil)

	status, err := s.grpcServer.GetServerStatus(context.Background(), &authv1.GetServerStatusRequest{})
	s.NoError(err)
	s.Equal("ok", status.Status)
	s.Equal(int32(42), status.TotalUsers)
}

func (s *GRPCServerTestSuite) TestGetServerLogs() {
	logs, err := s.grpcServer.GetServerLogs(context.Background(), &authv1.GetServerLogsRequest{})
	s.NoError(err)
	s.NotEmpty(logs.Logs)
}

func TestGRPCServerTestSuite(t *testing.T) {
	suite.Run(t, new(GRPCServerTestSuite))
}
