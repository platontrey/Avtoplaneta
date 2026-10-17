package repository

import (
	"context"

	"github.com/stretchr/testify/mock"

	"auth-service/db/sqlc"
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

func (m *MockUserRepository) UpdateProfile(id int64, name, username, inn string) (*domain.User, error) {
	args := m.Called(id, name, username, inn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) UpdatePassword(id int64, passwordHash string) error {
	args := m.Called(id, passwordHash)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateAvatar(id int64, avatarURL string) (*domain.User, error) {
	args := m.Called(id, avatarURL)
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

func (m *MockUserRepository) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) CountAll() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) CreateOutboxEvent(ctx context.Context, aggregateType string, aggregateID int64, eventType string, payload []byte) error {
	args := m.Called(ctx, aggregateType, aggregateID, eventType, payload)
	return args.Error(0)
}

func (m *MockUserRepository) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.OutboxEvent), args.Error(1)
}

func (m *MockUserRepository) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockUserRepository) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	args := m.Called(ctx, id, lastErr)
	return args.Error(0)
}

func (m *MockUserRepository) CleanupProcessedOutboxEvents(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
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
