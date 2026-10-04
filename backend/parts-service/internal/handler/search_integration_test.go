package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"parts-service/internal/catalog"
	"parts-service/internal/config"
	"parts-service/internal/domain"
	"parts-service/internal/search"
	"parts-service/internal/service"
)

type mockPartRepository struct {
	mock.Mock
}

func (m *mockPartRepository) Create(ctx context.Context, part *domain.Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}
func (m *mockPartRepository) FindByID(ctx context.Context, id int64) (*domain.Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}
func (m *mockPartRepository) FindAll(ctx context.Context) ([]domain.Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *mockPartRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *mockPartRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}
func (m *mockPartRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}
func (m *mockPartRepository) CreateBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error) {
	args := m.Called(ctx, parts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *mockPartRepository) MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error {
	args := m.Called(ctx, id, deleteAt)
	return args.Error(0)
}
func (m *mockPartRepository) BulkDelete(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}
func (m *mockPartRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	args := m.Called(ctx, updates)
	return args.Get(0).(int), args.Error(1)
}
func (m *mockPartRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	args := m.Called(ctx, supplierCode)
	return args.Get(0).(int64), args.Error(1)
}
func (m *mockPartRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}
func (m *mockPartRepository) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(domain.StatisticsResponse), args.Error(1)
}
func (m *mockPartRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}
func (m *mockPartRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	args := m.Called(ctx, amount)
	return args.Error(0)
}
func (m *mockPartRepository) DeleteExpiredParts(ctx context.Context, threshold time.Time) error {
	args := m.Called(ctx, threshold)
	return args.Error(0)
}
func (m *mockPartRepository) InventoryVersion(ctx context.Context) (string, error) {
	args := m.Called(ctx)
	return args.String(0), args.Error(1)
}
func (m *mockPartRepository) RenameSeller(ctx context.Context, sellerID int64, newName string) ([]domain.Part, error) {
	args := m.Called(ctx, sellerID, newName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *mockPartRepository) UpdatePartPhotos(ctx context.Context, id int64, photos domain.StringArray) error {
	args := m.Called(ctx, id, photos)
	return args.Error(0)
}
func (m *mockPartRepository) GetPartsForXML(ctx context.Context) ([]domain.Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]domain.Part), args.Error(1)
}
func (m *mockPartRepository) GetLastCreatedPart(ctx context.Context) (*domain.Part, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Part), args.Error(1)
}
func (m *mockPartRepository) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}
func (m *mockPartRepository) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}
func (m *mockPartRepository) GetNextSupplierCode(ctx context.Context) (string, error) {
	return "1", nil
}
func (m *mockPartRepository) PeekNextSupplierCode(ctx context.Context) (string, error) {
	return "1", nil
}

type mockElasticsearchClient struct {
	mock.Mock
}

func (m *mockElasticsearchClient) IndexPart(part *domain.Part) error {
	args := m.Called(part)
	return args.Error(0)
}
func (m *mockElasticsearchClient) DeletePartFromIndex(partID int64) error {
	args := m.Called(partID)
	return args.Error(0)
}
func (m *mockElasticsearchClient) SearchParts(query map[string]interface{}, from, size int) ([]search.ElasticsearchPart, int64, error) {
	args := m.Called(query, from, size)
	return args.Get(0).([]search.ElasticsearchPart), args.Get(1).(int64), args.Error(2)
}

func TestSearchIntegration_ElasticsearchSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockPartRepository)
	mockES := new(mockElasticsearchClient)
	cfg := &config.Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewInventoryService(mockRepo, mockES, cfg, nil)
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)
	h := NewHandler(svc, cat, nil, nil)

	router := gin.New()
	router.GET("/api/inventory", h.GetInventoryHandler)

	part1 := domain.Part{
		PartCore: domain.PartCore{
			ID:          10,
			Name:        "АКПП",
			Quantity:    2,
			Price:       45000.0,
			Category:    "Трансмиссия",
			Brand:       "Toyota",
			Model:       "Camry ACV30",
			Description: "Автоматическая коробка передач в сборе",
		},
	}

	mockES.On("SearchParts", mock.MatchedBy(func(query map[string]interface{}) bool {
		boolQ, ok := query["bool"].(map[string]interface{})
		if !ok {
			return false
		}
		must, ok := boolQ["must"].([]map[string]interface{})
		if !ok || len(must) != 2 {
			return false
		}
		return true
	}), 0, 20).Return([]search.ElasticsearchPart{
		{ID: part1.ID, Name: part1.Name},
	}, int64(1), nil)

	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		ids, ok := filters["ids_in"].([]int64)
		return ok && len(ids) == 1 && ids[0] == 10
	}), 0, 0).Return([]domain.Part{part1}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?search=АКПП+ACV30&page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, int64(10), response[0].ID)
	assert.Equal(t, "АКПП", response[0].Name)
	assert.Equal(t, "Camry ACV30", response[0].Model)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestSearchIntegration_ElasticsearchFallbackToDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockPartRepository)
	mockES := new(mockElasticsearchClient)
	cfg := &config.Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewInventoryService(mockRepo, mockES, cfg, nil)
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)
	h := NewHandler(svc, cat, nil, nil)

	router := gin.New()
	router.GET("/api/inventory", h.GetInventoryHandler)

	fallbackPart := domain.Part{
		PartCore: domain.PartCore{
			ID:       20,
			Name:     "Бампер передний",
			Quantity: 1,
			Price:    15000.0,
			Category: "Кузов снаружи",
			Brand:    "Honda",
			Model:    "Civic",
		},
	}

	mockES.On("SearchParts", mock.Anything, 0, 10).Return(
		[]search.ElasticsearchPart{},
		int64(0),
		errors.New("elasticsearch connection refused"),
	)

	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		searchVal, ok := filters["search"].(string)
		return ok && searchVal == "Бампер"
	}), 0, 10).Return([]domain.Part{fallbackPart}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?search=Бампер&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, int64(20), response[0].ID)
	assert.Equal(t, "Бампер передний", response[0].Name)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

func TestSearchIntegration_CategoryFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(mockPartRepository)
	mockES := new(mockElasticsearchClient)
	cfg := &config.Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	svc := service.NewInventoryService(mockRepo, mockES, cfg, nil)
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)
	h := NewHandler(svc, cat, nil, nil)

	router := gin.New()
	router.GET("/api/inventory", h.GetInventoryHandler)

	part := domain.Part{
		PartCore: domain.PartCore{
			ID:       30,
			Name:     "Тормозной диск",
			Quantity: 4,
			Price:    3500.0,
			Category: "Тормозная система",
			Brand:    "Brembo",
		},
	}

	mockES.On("SearchParts", mock.MatchedBy(func(query map[string]interface{}) bool {
		boolQ, ok := query["bool"].(map[string]interface{})
		if !ok {
			return false
		}
		filters, ok := boolQ["filter"].([]map[string]interface{})
		if !ok || len(filters) < 2 {
			return false
		}
		catFilter, ok := filters[1]["bool"].(map[string]interface{})
		if !ok {
			return false
		}
		shoulds, ok := catFilter["should"].([]map[string]interface{})
		if !ok {
			return false
		}
		for _, s := range shoulds {
			if termClause, ok := s["term"].(map[string]interface{}); ok {
				if val, ok := termClause["category"].(string); ok && val == "Тормозная система" {
					return true
				}
				if obj, ok := termClause["category"].(map[string]interface{}); ok && obj["value"] == "Тормозная система" {
					return true
				}
			}
		}
		return false
	}), 0, 50).Return([]search.ElasticsearchPart{
		{ID: part.ID, Name: part.Name},
	}, int64(1), nil)

	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		ids, ok := filters["ids_in"].([]int64)
		return ok && len(ids) == 1 && ids[0] == 30
	}), 0, 0).Return([]domain.Part{part}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?category=Тормозная+система", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []domain.Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "Тормозная система", response[0].Category)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
