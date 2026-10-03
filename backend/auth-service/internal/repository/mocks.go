package repository

import (
	"github.com/stretchr/testify/mock"

	"auth-service/internal/domain"
)

// MockUserRepository мок для UserRepository
type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByID(id int64) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmailOrName(identifier string) (*domain.User, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(id int64, params domain.UpdateUserParams) (*domain.User, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) Delete(id int64) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) FindAll() ([]domain.User, error) {
	args := m.Called()
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *MockUserRepository) ExistsByEmailOrName(email, name string) (bool, error) {
	args := m.Called(email, name)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CountAll() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

// MockActivityLogRepository мок для ActivityLogRepository
type MockActivityLogRepository struct {
	mock.Mock
}

func (m *MockActivityLogRepository) Create(log *domain.UserActivityLog) (*domain.UserActivityLog, error) {
	args := m.Called(log)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.UserActivityLog), args.Error(1)
}

func (m *MockActivityLogRepository) FindWithFilters(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error) {
	args := m.Called(filters)
	return args.Get(0).([]domain.UserActivityLog), args.Error(1)
}
