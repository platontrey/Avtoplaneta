package outbox

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"auth-service/db/sqlc"
	"auth-service/internal/domain"
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(user *domain.User) (*domain.User, error) {
	args := m.Called(user)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) FindByID(id int64) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) FindByEmailOrName(identifier string) (*domain.User, error) {
	args := m.Called(identifier)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) Update(id int64, params domain.UpdateUserParams) (*domain.User, error) {
	args := m.Called(id, params)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) Delete(id int64) error {
	return m.Called(id).Error(0)
}
func (m *MockUserRepo) FindAll() ([]domain.User, error) {
	args := m.Called()
	return args.Get(0).([]domain.User), args.Error(1)
}
func (m *MockUserRepo) ExistsByEmailOrName(email, name string) (bool, error) {
	args := m.Called(email, name)
	return args.Bool(0), args.Error(1)
}
func (m *MockUserRepo) ExistsByUsername(username string) (bool, error) {
	args := m.Called(username)
	return args.Bool(0), args.Error(1)
}
func (m *MockUserRepo) CountAll() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}
func (m *MockUserRepo) FindByEmail(email string) (*domain.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) UpdateProfile(id int64, name, username, inn string) (*domain.User, error) {
	args := m.Called(id, name, username, inn)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) UpdatePassword(id int64, passwordHash string) error {
	return m.Called(id, passwordHash).Error(0)
}
func (m *MockUserRepo) UpdateAvatar(id int64, avatarURL string) (*domain.User, error) {
	args := m.Called(id, avatarURL)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}
func (m *MockUserRepo) CreateOutboxEvent(ctx context.Context, aggregateType string, aggregateID int64, eventType string, payload []byte) error {
	return m.Called(ctx, aggregateType, aggregateID, eventType, payload).Error(0)
}
func (m *MockUserRepo) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.OutboxEvent), args.Error(1)
}
func (m *MockUserRepo) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockUserRepo) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return m.Called(ctx, id, lastErr).Error(0)
}
func (m *MockUserRepo) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

func TestDispatcher_ProcessSellerRenamed_Success(t *testing.T) {
	mockRepo := new(MockUserRepo)

	dispatcher := NewDispatcher(nil, mockRepo, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	payload, _ := json.Marshal(map[string]interface{}{"seller_id": 42, "name": "ИП Автомир"})
	event := sqlc.OutboxEvent{
		ID:            1,
		AggregateType: "user",
		AggregateID:   42,
		EventType:     "seller_renamed",
		Payload:       payload,
		Status:        "pending",
	}

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(1)).Return(nil)
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
