package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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

func TestGetItem_SuccessAndETag(t *testing.T) {
	svc := new(mockItemService)
	now := time.Now()
	item := &domain.Item{
		ID:        1,
		Name:      "Тестовый товар",
		UpdatedAt: now,
	}
	svc.On("GetItem", mock.Anything, int64(1)).Return(item, nil)

	router := setupTestRouter(svc)

	// 1. Первый запрос без ETag -> 200 OK и заголовок ETag
	w1 := httptest.NewRecorder()
	req1, _ := http.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusOK, w1.Code)
	etag := w1.Header().Get("ETag")
	require.NotEmpty(t, etag)

	// 2. Повторный запрос с If-None-Match -> 304 Not Modified
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest(http.MethodGet, "/api/v1/items/1", nil)
	req2.Header.Set("If-None-Match", etag)
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusNotModified, w2.Code)
	require.Empty(t, w2.Body.String())
}
