package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"template-service/internal/domain"
)

type mockItemService struct {
	mock.Mock
}

func (m *mockItemService) GetItem(ctx context.Context, id int64) (*domain.Item, error) {
	args := m.Called(ctx, id)
	if item := args.Get(0); item != nil {
		return item.(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockItemService) ListItems(ctx context.Context, page, limit int32) ([]domain.Item, error) {
	args := m.Called(ctx, page, limit)
	return args.Get(0).([]domain.Item), args.Error(1)
}

func (m *mockItemService) CreateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	args := m.Called(ctx, item)
	if it := args.Get(0); it != nil {
		return it.(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockItemService) UpdateItem(ctx context.Context, item *domain.Item) (*domain.Item, error) {
	args := m.Called(ctx, item)
	if it := args.Get(0); it != nil {
		return it.(*domain.Item), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockItemService) DeleteItem(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter(svc *mockItemService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewHandler(svc)
	SetupRoutes(r, h)
	return r
}

func TestHealthEndpoint(t *testing.T) {
	svc := new(mockItemService)
	router := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), `"status":"healthy"`)
	require.Contains(t, w.Body.String(), `"service":"template-service"`)
}

func TestMetricsEndpoint(t *testing.T) {
	svc := new(mockItemService)
	router := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
	require.Contains(t, w.Body.String(), "template_service_http_requests_total")
}

func TestGetItem_NotFound(t *testing.T) {
	svc := new(mockItemService)
	svc.On("GetItem", mock.Anything, int64(999)).Return(nil, domain.ErrItemNotFound)

	router := setupTestRouter(svc)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/items/999", nil)
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusNotFound, w.Code)
	svc.AssertExpectations(t)
}
