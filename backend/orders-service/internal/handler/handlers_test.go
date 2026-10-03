package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"orders-service/internal/domain"
)

type MockOrdersService struct {
	mock.Mock
}

func (m *MockOrdersService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrdersService) GetCompletedOrders(ctx context.Context) ([]domain.Order, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Order), args.Error(1)
}

func (m *MockOrdersService) CreateOrder(ctx context.Context, req domain.CreateOrderRequest, userID int64, userName string) (*domain.Order, error) {
	args := m.Called(ctx, req, userID, userName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrdersService) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	args := m.Called(ctx, orderID, status)
	return args.Error(0)
}

func (m *MockOrdersService) UpdateOrderDetails(ctx context.Context, orderID int64, req domain.UpdateOrderDetailsRequest) (*domain.Order, error) {
	args := m.Called(ctx, orderID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Order), args.Error(1)
}

func (m *MockOrdersService) CompleteOrder(ctx context.Context, orderID int64) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockOrdersService) DeleteOrder(ctx context.Context, orderID int64) error {
	args := m.Called(ctx, orderID)
	return args.Error(0)
}

func (m *MockOrdersService) AddOrderItem(ctx context.Context, orderID int64, req domain.AddOrderItemRequest) error {
	args := m.Called(ctx, orderID, req)
	return args.Error(0)
}

func (m *MockOrdersService) UpdateOrderItem(ctx context.Context, orderID, itemID int64, req domain.UpdateOrderItemRequest) error {
	args := m.Called(ctx, orderID, itemID, req)
	return args.Error(0)
}

func (m *MockOrdersService) DeleteOrderItem(ctx context.Context, orderID, itemID int64) error {
	args := m.Called(ctx, orderID, itemID)
	return args.Error(0)
}

func (m *MockOrdersService) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.MonthlySales), args.Error(1)
}

func (m *MockOrdersService) ListCustomers(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error) {
	args := m.Called(ctx, category, search, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.CustomerWithStats), args.Error(1)
}

func (m *MockOrdersService) GetCustomer(ctx context.Context, id int64) (*domain.CustomerDetails, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.CustomerDetails), args.Error(1)
}

func (m *MockOrdersService) CreateCustomer(ctx context.Context, req domain.CreateCustomerRequest) (*domain.Customer, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockOrdersService) UpdateCustomer(ctx context.Context, id int64, req domain.UpdateCustomerRequest) (*domain.Customer, error) {
	args := m.Called(ctx, id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Customer), args.Error(1)
}

func (m *MockOrdersService) DeleteCustomer(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

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

func setupTestRouter(svc *MockOrdersService, pub *MockEventPublisher) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc, pub)
	SetupRoutes(r, h)
	return r
}

func TestHandler_AuthMiddleware(t *testing.T) {
	mockSvc := new(MockOrdersService)
	mockPub := new(MockEventPublisher)
	router := setupTestRouter(mockSvc, mockPub)

	t.Run("Unauthorized Without Headers", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/orders", nil)
		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusUnauthorized, resp.Code)
	})

	t.Run("Authorized With ForwardAuth Headers", func(t *testing.T) {
		mockSvc.On("GetOrders", mock.Anything).Return([]domain.Order{}, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/orders", nil)
		req.Header.Set("X-User-ID", "1")
		req.Header.Set("X-User-Email", "test@test.com")
		req.Header.Set("X-User-Name", "Admin")
		req.Header.Set("X-User-Role", "admin")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		mockSvc.AssertExpectations(t)
	})
}

func TestHandler_CreateOrder(t *testing.T) {
	mockSvc := new(MockOrdersService)
	mockPub := new(MockEventPublisher)
	router := setupTestRouter(mockSvc, mockPub)

	expectedOrder := &domain.Order{
		ID:          1,
		BuyerNumber: "+79991234567",
		TotalAmount: 15000,
	}

	createReq := domain.CreateOrderRequest{
		PartID:      50,
		BuyerNumber: "+79991234567",
		Items: []domain.CreateOrderItemInput{
			{PartID: 50, Quantity: 1},
		},
	}

	mockSvc.On("CreateOrder", mock.Anything, createReq, int64(1), "Admin").Return(expectedOrder, nil)
	mockPub.On("PublishUserAction", mock.Anything, "1", "create_order", mock.Anything).Return(nil)

	body, _ := json.Marshal(createReq)
	req, _ := http.NewRequest(http.MethodPost, "/orders", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", "1")
	req.Header.Set("X-User-Name", "Admin")

	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusCreated, resp.Code)

	var res domain.Order
	err := json.Unmarshal(resp.Body.Bytes(), &res)
	require.NoError(t, err)
	assert.Equal(t, int64(1), res.ID)
	assert.Equal(t, 15000.0, res.TotalAmount)

	mockSvc.AssertExpectations(t)
	mockPub.AssertExpectations(t)
}

func TestHandler_CustomerEndpoints(t *testing.T) {
	mockSvc := new(MockOrdersService)
	mockPub := new(MockEventPublisher)
	router := setupTestRouter(mockSvc, mockPub)

	t.Run("List Customers", func(t *testing.T) {
		mockSvc.On("ListCustomers", mock.Anything, "", "", int32(50), int32(0)).
			Return([]domain.CustomerWithStats{
				{Customer: domain.Customer{ID: 10, Name: "ООО Спецтех"}},
			}, nil).Once()

		req, _ := http.NewRequest(http.MethodGet, "/customers", nil)
		req.Header.Set("X-User-ID", "1")

		resp := httptest.NewRecorder()
		router.ServeHTTP(resp, req)

		assert.Equal(t, http.StatusOK, resp.Code)
		assert.Contains(t, resp.Body.String(), "ООО Спецтех")
		mockSvc.AssertExpectations(t)
	})
}
