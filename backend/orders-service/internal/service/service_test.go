package service

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"orders-service/internal/domain"
)

// MockOrderRepository мок для OrderRepository
type MockOrderRepository struct {
	mock.Mock
}

func (m *MockOrderRepository) Create(ctx context.Context, order *domain.Order) error {
	args := m.Called(ctx, order)
	if order.ID == 0 {
		order.ID = 100
	}
	return args.Error(0)
}

func (m *MockOrderRepository) CreateItem(ctx context.Context, item *domain.OrderItem) error {
	args := m.Called(ctx, item)
	if item.ID == 0 {
		item.ID = 200
	}
	return args.Error(0)
}

func (m *MockOrderRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindWithItemsByID(ctx context.Context, id int64) (*domain.Order, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindAll(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindActive(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindCompleted(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindWithItems(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrderRepository) FindOrderItem(ctx context.Context, orderID, partID int64) (*domain.OrderItem, error) {
	args := m.Called(ctx, orderID, partID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.OrderItem), args.Error(1)
}

func (m *MockOrderRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	args := m.Called(ctx, id, status)
	return args.Error(0)
}

func (m *MockOrderRepository) CompleteRecord(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateItem(ctx context.Context, item *domain.OrderItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockOrderRepository) DeleteItem(ctx context.Context, orderID, itemID int64) error {
	args := m.Called(ctx, orderID, itemID)
	return args.Error(0)
}

func (m *MockOrderRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) DeleteItemsByOrderID(ctx context.Context, orderID int64) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockOrderRepository) MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockOrderRepository) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.MonthlySales), args.Error(1)
}

func (m *MockOrderRepository) CreateSalesHistory(ctx context.Context, history *domain.SalesHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateSellerName(ctx context.Context, sellerID int64, name string) error {
	args := m.Called(ctx, sellerID, name)
	return args.Error(0)
}

func (m *MockOrderRepository) GetPool() *pgxpool.Pool {
	return nil
}

// Customer repository mocks
func (m *MockOrderRepository) ListCustomersWithStats(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error) {
	args := m.Called(ctx, category, search, limit, offset)
	return args.Get(0).([]domain.CustomerWithStats), args.Error(1)
}

func (m *MockOrderRepository) GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockOrderRepository) GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockOrderRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	if customer.ID == 0 {
		customer.ID = 10
	}
	return args.Error(0)
}

func (m *MockOrderRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	args := m.Called(ctx, customer)
	return args.Error(0)
}

func (m *MockOrderRepository) DeleteCustomer(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockOrderRepository) GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]domain.Order, error) {
	args := m.Called(ctx, customerID)
	return args.Get(0).([]domain.Order), args.Error(1)
}

// MockPartRepositoryForOrders мок для PartRepositoryForOrders
type MockPartRepositoryForOrders struct {
	mock.Mock
}

func (m *MockPartRepositoryForOrders) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}

func (m *MockPartRepositoryForOrders) UpdateQuantity(ctx context.Context, id int64, newQuantity int) error {
	args := m.Called(ctx, id, newQuantity)
	return args.Error(0)
}

func (m *MockPartRepositoryForOrders) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockPartRepositoryForOrders) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockPartRepositoryForOrders) DeletePart(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockCacheService мок для CacheService
type MockCacheService struct {
	mock.Mock
}

func (m *MockCacheService) GetOrders() ([]domain.Order, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockCacheService) SetOrders(orders []domain.Order) error {
	args := m.Called(orders)
	return args.Error(0)
}

func (m *MockCacheService) InvalidateOrders() error {
	args := m.Called()
	return args.Error(0)
}

// MockEventPublisher мок для EventPublisher
type MockEventPublisher struct {
	mock.Mock
}

func (m *MockEventPublisher) PublishOrderCompleted(ctx context.Context, orderID int64, amount float64) error {
	args := m.Called(ctx, orderID, amount)
	return args.Error(0)
}

func (m *MockEventPublisher) PublishUserAction(ctx context.Context, userID, action string, details map[string]interface{}) error {
	args := m.Called(ctx, userID, action, details)
	return args.Error(0)
}

// Тест создания заказа на деталь с привязкой клиента
func TestCreateOrder_SingleItem(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepositoryForOrders)
	mockCache := new(MockCacheService)
	mockPublisher := new(MockEventPublisher)

	svc := NewOrdersService(mockOrderRepo, mockPartRepo, mockCache, mockPublisher)
	ctx := context.Background()

	part := &domain.Part{
		ID:       50,
		Name:     "Генератор Bosch",
		Price:    15000,
		Quantity: 3,
		Location: "Склад 1",
	}

	mockPartRepo.On("FindByID", ctx, int64(50)).Return(part, nil)
	mockOrderRepo.On("GetCustomerByPhone", ctx, "+79991234567").Return((*domain.Customer)(nil), nil)
	mockOrderRepo.On("CreateCustomer", ctx, mock.AnythingOfType("*domain.Customer")).Return(nil)
	mockOrderRepo.On("Create", ctx, mock.AnythingOfType("*domain.Order")).Return(nil)
	mockOrderRepo.On("CreateItem", ctx, mock.AnythingOfType("*domain.OrderItem")).Return(nil)
	mockOrderRepo.On("FindWithItemsByID", mock.Anything, int64(100)).Return(&domain.Order{
		ID:          100,
		TotalAmount: 15000,
		Status:      "Ожидает предоплаты",
		PartID:      50,
		Seller:      "Менеджер Иван",
		Items: []domain.OrderItem{
			{PartID: 50, Quantity: 1, Price: 15000},
		},
	}, nil)
	mockCache.On("InvalidateOrders").Return(nil)

	req := domain.CreateOrderRequest{
		PartID:      50,
		BuyerNumber: "+79991234567",
		Source:      "drom",
		Items: []domain.CreateOrderItemInput{
			{PartID: 50, Quantity: 1},
		},
	}

	created, err := svc.CreateOrder(ctx, req, 1, "Менеджер Иван")
	require.NoError(t, err)
	assert.NotNil(t, created)
	assert.Equal(t, 15000.0, created.TotalAmount)
	assert.Equal(t, "Ожидает предоплаты", created.Status)
	assert.Equal(t, int64(50), created.PartID)
	assert.Equal(t, "Менеджер Иван", created.Seller)

	mockOrderRepo.AssertExpectations(t)
	mockPartRepo.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

// Тест валидации: заказ без деталей запрещен
func TestCreateOrder_EmptyItems(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepositoryForOrders)
	mockCache := new(MockCacheService)
	mockPublisher := new(MockEventPublisher)

	svc := NewOrdersService(mockOrderRepo, mockPartRepo, mockCache, mockPublisher)
	ctx := context.Background()

	req := domain.CreateOrderRequest{
		BuyerNumber: "+79991234567",
		Items:       []domain.CreateOrderItemInput{},
	}

	_, err := svc.CreateOrder(ctx, req, 1, "Менеджер")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "at least one part must be selected")
}

// Тест смены статуса заказа
func TestUpdateOrderStatus(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepositoryForOrders)
	mockCache := new(MockCacheService)
	mockPublisher := new(MockEventPublisher)

	svc := NewOrdersService(mockOrderRepo, mockPartRepo, mockCache, mockPublisher)
	ctx := context.Background()

	t.Run("ValidStatus", func(t *testing.T) {
		mockOrderRepo.On("Update", ctx, int64(200), mock.MatchedBy(func(u map[string]interface{}) bool {
			return u["status"] == "Готов к выдаче"
		})).Return(nil)
		mockCache.On("InvalidateOrders").Return(nil)

		err := svc.UpdateOrderStatus(ctx, 200, "Готов к выдаче")
		require.NoError(t, err)
	})

	t.Run("InvalidStatus", func(t *testing.T) {
		err := svc.UpdateOrderStatus(ctx, 200, "НеизвестныйСтатус")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid status")
	})
}

// Тест CRUD операций с клиентами CRM
func TestCustomerOperations(t *testing.T) {
	mockOrderRepo := new(MockOrderRepository)
	mockPartRepo := new(MockPartRepositoryForOrders)
	mockCache := new(MockCacheService)
	mockPublisher := new(MockEventPublisher)

	svc := NewOrdersService(mockOrderRepo, mockPartRepo, mockCache, mockPublisher)
	ctx := context.Background()

	t.Run("CreateCustomer", func(t *testing.T) {
		mockOrderRepo.On("GetCustomerByPhone", ctx, "+79998887766").Return((*domain.Customer)(nil), nil)
		mockOrderRepo.On("CreateCustomer", ctx, mock.MatchedBy(func(c *domain.Customer) bool {
			return c.Name == "ООO Автомир" && c.Phone == "+79998887766" && c.Category == "b2b"
		})).Return(nil)

		req := domain.CreateCustomerRequest{
			Name:            "ООO Автомир",
			Phone:           "+79998887766",
			Category:        "b2b",
			DiscountPercent: 10,
		}

		c, err := svc.CreateCustomer(ctx, req)
		require.NoError(t, err)
		assert.Equal(t, "ООO Автомир", c.Name)
	})

	t.Run("GetCustomerDetails", func(t *testing.T) {
		customer := &domain.Customer{ID: 10, Name: "ООO Автомир", Phone: "+79998887766"}
		orders := []domain.Order{
			{
				ID:          1,
				OrderNumber: "ORD-001",
				TotalAmount: 10000,
				Status:      "green",
				Items: []domain.OrderItem{
					{PartID: 1, Quantity: 1, Price: 10000},
				},
			},
		}

		mockOrderRepo.On("GetCustomerByID", ctx, int64(10)).Return(customer, nil)
		mockOrderRepo.On("GetOrdersByCustomerID", ctx, int64(10)).Return(orders, nil)

		details, err := svc.GetCustomer(ctx, 10)
		require.NoError(t, err)
		assert.Equal(t, "ООO Автомир", details.Name)
		assert.Equal(t, int64(1), details.TotalOrders)
		assert.Equal(t, 10000.0, details.TotalSpent)
		assert.Len(t, details.Orders, 1)
	})

	t.Run("DeleteCustomer", func(t *testing.T) {
		customer := &domain.Customer{ID: 15, Name: "Клиент на удаление"}
		mockOrderRepo.On("GetCustomerByID", ctx, int64(15)).Return(customer, nil)
		mockOrderRepo.On("DeleteCustomer", ctx, int64(15)).Return(nil)

		err := svc.DeleteCustomer(ctx, 15)
		require.NoError(t, err)
	})
}
