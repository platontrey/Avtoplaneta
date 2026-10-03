package service

import (
	"github.com/stretchr/testify/mock"

	"auth-service/internal/domain"
)

// MockAuthService мок для AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthenticateUser(email, password string) (*domain.User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) CreateUserFromGoogle(email, name string) (*domain.User, error) {
	args := m.Called(email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GetCurrentUser(userID int64) (*domain.User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) Logout(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) CreateUser(req domain.CreateUserRequest) (*domain.User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) GetUsers() ([]domain.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockAuthService) UpdateUser(userID int64, req domain.UpdateUserRequest) (*domain.User, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockAuthService) DeleteUser(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) GetTotalUsersCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockAuthService) LogUserActivity(user *domain.User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error {
	args := m.Called(user, action, resourceType, resourceID, details, ip, userAgent)
	return args.Error(0)
}

func (m *MockAuthService) GetUserActivityLogs(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error) {
	args := m.Called(filters)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.UserActivityLog), args.Error(1)
}

func (m *MockAuthService) GenerateCSRFToken(userID *int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}
