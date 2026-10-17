package outbox

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	partsv1 "avtoplaneta/gen/parts/v1"
	"parts-service/db/sqlc"
	"parts-service/internal/domain"
	"parts-service/internal/search"
)

// MockRepository реализует repository.PartRepository для тестов
type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Create(ctx context.Context, part *domain.Part) error {
	return m.Called(ctx, part).Error(0)
}
func (m *MockRepository) CreateBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error) {
	args := m.Called(ctx, parts)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}
func (m *MockRepository) FindAll(ctx context.Context) ([]domain.Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return m.Called(ctx, id, updates).Error(0)
}
func (m *MockRepository) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockRepository) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	return m.Called(ctx, id, amount, operationID).Error(0)
}
func (m *MockRepository) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	return m.Called(ctx, id, amount, operationID).Error(0)
}
func (m *MockRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error {
	return m.Called(ctx, id, deleteAt).Error(0)
}
func (m *MockRepository) DeleteExpiredParts(ctx context.Context, before time.Time) error {
	return m.Called(ctx, before).Error(0)
}
func (m *MockRepository) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.StatisticsResponse), args.Error(1)
}
func (m *MockRepository) InventoryVersion(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *MockRepository) RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error) {
	args := m.Called(ctx, sellerID, name)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) BulkDelete(ctx context.Context, ids []int64) error {
	return m.Called(ctx, ids).Error(0)
}
func (m *MockRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	args := m.Called(ctx, updates)
	return args.Int(0), args.Error(1)
}
func (m *MockRepository) BulkUpdateWithProto(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error) {
	args := m.Called(ctx, req, filterMap)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	args := m.Called(ctx, supplierCode)
	return args.Get(0).(int64), args.Error(1)
}
func (m *MockRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}
func (m *MockRepository) GetNextSupplierCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *MockRepository) PeekNextSupplierCode(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *MockRepository) GetNextPartNumber(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *MockRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}
func (m *MockRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	return m.Called(ctx, amount).Error(0)
}
func (m *MockRepository) UpdatePartPhotos(ctx context.Context, id int64, photos domain.StringArray) error {
	return m.Called(ctx, id, photos).Error(0)
}
func (m *MockRepository) GetPartsForXML(ctx context.Context) ([]domain.Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *MockRepository) GetLastCreatedPart(ctx context.Context) (*domain.Part, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}

func (m *MockRepository) CountWithFilters(ctx context.Context, filters map[string]interface{}) (int64, error) {
	args := m.Called(ctx, filters)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockRepository) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.OutboxEvent), args.Error(1)
}

func (m *MockRepository) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

func (m *MockRepository) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return m.Called(ctx, id, lastErr).Error(0)
}

func (m *MockRepository) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

// MockESClient реализует search.ElasticsearchClient для тестов
type MockESClient struct {
	mock.Mock
}

func (m *MockESClient) IndexPart(part *domain.Part) error {
	return m.Called(part).Error(0)
}

func (m *MockESClient) DeletePartFromIndex(partID int64) error {
	return m.Called(partID).Error(0)
}

func (m *MockESClient) SearchParts(query map[string]interface{}, from, size int) ([]search.ElasticsearchPart, int64, error) {
	args := m.Called(query, from, size)
	return args.Get(0).([]search.ElasticsearchPart), args.Get(1).(int64), args.Error(2)
}

func TestDispatcher_ProcessEvent_PartCreated_Success(t *testing.T) {
	mockRepo := new(MockRepository)
	mockES := new(MockESClient)

	dispatcher := NewDispatcher(nil, mockRepo, mockES, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testPart := &domain.Part{
		PartCore: domain.PartCore{
			ID:       101,
			Name:     "Тестовая фара",
			Quantity: 1,
			Photos:   domain.StringArray{"/uploads/101/photo.jpg"},
		},
	}

	event := sqlc.OutboxEvent{
		ID:            1,
		AggregateType: "part",
		AggregateID:   101,
		EventType:     "part_created",
		Status:        "pending",
	}

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockRepo.On("FindByID", mock.Anything, int64(101)).Return(testPart, nil)
	mockES.On("IndexPart", testPart).Return(nil)
	mockRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(1)).Return(nil)

	// Следующий вызов вернет пустой список, чтобы цикл не зациклился
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockES.AssertExpectations(t)
}

func TestDispatcher_ProcessEvent_PartDeleted(t *testing.T) {
	mockRepo := new(MockRepository)
	mockES := new(MockESClient)

	dispatcher := NewDispatcher(nil, mockRepo, mockES, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := sqlc.OutboxEvent{
		ID:            2,
		AggregateType: "part",
		AggregateID:   102,
		EventType:     "part_deleted",
		Status:        "pending",
	}

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockES.On("DeletePartFromIndex", int64(102)).Return(nil)
	mockRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(2)).Return(nil)
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockES.AssertExpectations(t)
}

func TestDispatcher_ProcessEvent_ESFailure_MarksFailed(t *testing.T) {
	mockRepo := new(MockRepository)
	mockES := new(MockESClient)

	dispatcher := NewDispatcher(nil, mockRepo, mockES, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testPart := &domain.Part{
		PartCore: domain.PartCore{
			ID:       103,
			Name:     "Капот",
			Quantity: 1,
		},
	}

	event := sqlc.OutboxEvent{
		ID:            3,
		AggregateType: "part",
		AggregateID:   103,
		EventType:     "part_updated",
		Status:        "pending",
	}

	esErr := errors.New("connection timeout to elasticsearch")

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockRepo.On("FindByID", mock.Anything, int64(103)).Return(testPart, nil)
	mockES.On("IndexPart", testPart).Return(esErr)
	mockRepo.On("MarkOutboxEventFailed", mock.Anything, int64(3), mock.MatchedBy(func(errStr string) bool {
		return assert.Contains(t, errStr, "failed to index part 103 in Elasticsearch")
	})).Return(nil)
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockES.AssertExpectations(t)
}

func TestDispatcher_ProcessEvent_PartNotFoundInDB(t *testing.T) {
	mockRepo := new(MockRepository)
	mockES := new(MockESClient)

	dispatcher := NewDispatcher(nil, mockRepo, mockES, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := sqlc.OutboxEvent{
		ID:            4,
		AggregateType: "part",
		AggregateID:   104,
		EventType:     "part_photos_updated",
		Status:        "pending",
	}

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockRepo.On("FindByID", mock.Anything, int64(104)).Return(nil, pgx.ErrNoRows)
	mockES.On("DeletePartFromIndex", int64(104)).Return(nil)
	mockRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(4)).Return(nil)
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
	mockES.AssertExpectations(t)
}

func TestDispatcher_ProcessEvent_UnknownEventType(t *testing.T) {
	mockRepo := new(MockRepository)
	mockES := new(MockESClient)

	dispatcher := NewDispatcher(nil, mockRepo, mockES, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	event := sqlc.OutboxEvent{
		ID:            5,
		AggregateType: "part",
		AggregateID:   105,
		EventType:     "unknown_future_event",
		Status:        "pending",
	}

	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(5)).Return(nil)
	mockRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}
