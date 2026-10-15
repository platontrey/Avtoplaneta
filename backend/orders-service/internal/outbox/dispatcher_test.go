package outbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"orders-service/db/sqlc"
	"orders-service/internal/domain"
)

type MockOrderRepo struct {
	mock.Mock
}

func (m *MockOrderRepo) Create(ctx context.Context, order *domain.Order) error {
	return m.Called(ctx, order).Error(0)
}
func (m *MockOrderRepo) CreateItem(ctx context.Context, item *domain.OrderItem) error {
	return m.Called(ctx, item).Error(0)
}
func (m *MockOrderRepo) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindWithItemsByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindAll(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindActive(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindCompleted(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindWithItems(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockOrderRepo) FindOrderItem(ctx context.Context, orderID, partID int64) (*domain.OrderItem, error) {
	args := m.Called(ctx, orderID, partID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrderItem), args.Error(1)
}
func (m *MockOrderRepo) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	return m.Called(ctx, id, updates).Error(0)
}
func (m *MockOrderRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	return m.Called(ctx, id, status).Error(0)
}
func (m *MockOrderRepo) CompleteRecord(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockOrderRepo) UpdateItem(ctx context.Context, item *domain.OrderItem) error {
	return m.Called(ctx, item).Error(0)
}
func (m *MockOrderRepo) DeleteItem(ctx context.Context, orderID, itemID int64) error {
	return m.Called(ctx, orderID, itemID).Error(0)
}
func (m *MockOrderRepo) Delete(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockOrderRepo) DeleteItemsByOrderID(ctx context.Context, orderID int64) error {
	return m.Called(ctx, orderID).Error(0)
}
func (m *MockOrderRepo) MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error {
	return m.Called(ctx, before).Error(0)
}
func (m *MockOrderRepo) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.MonthlySales), args.Error(1)
}
func (m *MockOrderRepo) CreateSalesHistory(ctx context.Context, history *domain.SalesHistory) error {
	return m.Called(ctx, history).Error(0)
}
func (m *MockOrderRepo) UpdateSellerName(ctx context.Context, sellerID int64, name string) error {
	return m.Called(ctx, sellerID, name).Error(0)
}
func (m *MockOrderRepo) GetPool() *pgxpool.Pool {
	return nil
}
func (m *MockOrderRepo) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	return m.Called(ctx, customer).Error(0)
}
func (m *MockOrderRepo) GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}
func (m *MockOrderRepo) GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}
func (m *MockOrderRepo) ListCustomersWithStats(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error) {
	args := m.Called(ctx, category, search, limit, offset)
	return args.Get(0).([]domain.CustomerWithStats), args.Error(1)
}
func (m *MockOrderRepo) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	return m.Called(ctx, customer).Error(0)
}
func (m *MockOrderRepo) DeleteCustomer(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockOrderRepo) GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]domain.Order, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockOrderRepo) CreateOutboxEvent(ctx context.Context, aggregateType string, aggregateID int64, eventType string, payload []byte) error {
	return m.Called(ctx, aggregateType, aggregateID, eventType, payload).Error(0)
}
func (m *MockOrderRepo) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]sqlc.OutboxEvent), args.Error(1)
}
func (m *MockOrderRepo) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}
func (m *MockOrderRepo) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return m.Called(ctx, id, lastErr).Error(0)
}
func (m *MockOrderRepo) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return m.Called(ctx).Error(0)
}

type MockPartRepoForOrders struct {
	mock.Mock
}

func (m *MockPartRepoForOrders) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}
func (m *MockPartRepoForOrders) UpdateQuantity(ctx context.Context, id int64, newQuantity int) error {
	return m.Called(ctx, id, newQuantity).Error(0)
}
func (m *MockPartRepoForOrders) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	return m.Called(ctx, id, amount, operationID).Error(0)
}
func (m *MockPartRepoForOrders) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	return m.Called(ctx, id, amount, operationID).Error(0)
}
func (m *MockPartRepoForOrders) DeletePart(ctx context.Context, id int64) error {
	return m.Called(ctx, id).Error(0)
}

type MockCache struct {
	mock.Mock
}

func (m *MockCache) GetOrders() ([]domain.Order, error) {
	args := m.Called()
	return args.Get(0).([]domain.Order), args.Error(1)
}
func (m *MockCache) SetOrders(orders []domain.Order) error {
	return m.Called(orders).Error(0)
}
func (m *MockCache) InvalidateOrders() error {
	return m.Called().Error(0)
}

func TestDispatcher_ProcessOrderCompleted_Success(t *testing.T) {
	mockOrderRepo := new(MockOrderRepo)
	mockPartRepo := new(MockPartRepoForOrders)
	mockCache := new(MockCache)

	dispatcher := NewDispatcher(nil, mockOrderRepo, mockPartRepo, nil, mockCache)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testOrder := &domain.Order{
		ID:       55,
		Discount: 0,
		Items: []domain.OrderItem{
			{PartID: 201, Quantity: 1, Price: 1000},
		},
	}

	testPart := &domain.Part{
		ID:       201,
		Quantity: 3,
	}

	payload, _ := json.Marshal(map[string]interface{}{"order_id": 55, "amount": 1000.0})
	event := sqlc.OutboxEvent{
		ID:            1,
		AggregateType: "order",
		AggregateID:   55,
		EventType:     "order_completed",
		Payload:       payload,
		Status:        "pending",
	}

	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockOrderRepo.On("FindWithItemsByID", mock.Anything, int64(55)).Return(testOrder, nil)
	mockPartRepo.On("FindByID", mock.Anything, int64(201)).Return(testPart, nil)
	mockPartRepo.On("DecreaseQuantity", mock.Anything, int64(201), 1, "complete-order-55-part-201").Return(nil)
	mockCache.On("InvalidateOrders").Return(nil)
	mockOrderRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(1)).Return(nil)
	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockOrderRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestDispatcher_ProcessOrderCompleted_DepletedPart(t *testing.T) {
	mockOrderRepo := new(MockOrderRepo)
	mockPartRepo := new(MockPartRepoForOrders)
	mockCache := new(MockCache)

	dispatcher := NewDispatcher(nil, mockOrderRepo, mockPartRepo, nil, mockCache)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testOrder := &domain.Order{
		ID:       56,
		Discount: 0,
		Items: []domain.OrderItem{
			{PartID: 202, Quantity: 1, Price: 500},
		},
	}

	testPart := &domain.Part{
		ID:       202,
		Quantity: 1, // Will be depleted
	}

	payload, _ := json.Marshal(map[string]interface{}{"order_id": 56, "amount": 500.0})
	event := sqlc.OutboxEvent{
		ID:            2,
		AggregateType: "order",
		AggregateID:   56,
		EventType:     "order_completed",
		Payload:       payload,
		Status:        "pending",
	}

	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockOrderRepo.On("FindWithItemsByID", mock.Anything, int64(56)).Return(testOrder, nil)
	mockPartRepo.On("FindByID", mock.Anything, int64(202)).Return(testPart, nil)
	mockPartRepo.On("DeletePart", mock.Anything, int64(202)).Return(nil)
	mockCache.On("InvalidateOrders").Return(nil)
	mockOrderRepo.On("MarkOutboxEventProcessed", mock.Anything, int64(2)).Return(nil)
	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockOrderRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
}

func TestDispatcher_ProcessOrderCompleted_PartsServiceFailure(t *testing.T) {
	mockOrderRepo := new(MockOrderRepo)
	mockPartRepo := new(MockPartRepoForOrders)

	dispatcher := NewDispatcher(nil, mockOrderRepo, mockPartRepo, nil, nil)
	dispatcher.pollInterval = 10 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	testOrder := &domain.Order{
		ID: 57,
		Items: []domain.OrderItem{
			{PartID: 203, Quantity: 2, Price: 1500},
		},
	}

	testPart := &domain.Part{
		ID:       203,
		Quantity: 5,
	}

	payload, _ := json.Marshal(map[string]interface{}{"order_id": 57, "amount": 3000.0})
	event := sqlc.OutboxEvent{
		ID:            3,
		AggregateType: "order",
		AggregateID:   57,
		EventType:     "order_completed",
		Payload:       payload,
		Status:        "pending",
	}

	gRPCErr := errors.New("gRPC connection error to parts-service")

	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{event}, nil).Once()
	mockOrderRepo.On("FindWithItemsByID", mock.Anything, int64(57)).Return(testOrder, nil)
	mockPartRepo.On("FindByID", mock.Anything, int64(203)).Return(testPart, nil)
	mockPartRepo.On("DecreaseQuantity", mock.Anything, int64(203), 2, "complete-order-57-part-203").Return(gRPCErr)
	mockOrderRepo.On("MarkOutboxEventFailed", mock.Anything, int64(3), mock.MatchedBy(func(errStr string) bool {
		return assert.Contains(t, errStr, "failed to decrease part 203 quantity")
	})).Return(nil)
	mockOrderRepo.On("FetchPendingOutboxEvents", mock.Anything, int32(50)).Return([]sqlc.OutboxEvent{}, nil)

	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	err := dispatcher.Start(ctx)
	assert.NoError(t, err)

	mockOrderRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
}
