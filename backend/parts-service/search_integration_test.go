package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// TestSearchIntegration_ElasticsearchSuccess тестирует полный путь поиска через HTTP-хэндлер с успешным Elasticsearch
func TestSearchIntegration_ElasticsearchSuccess(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockPartRepository)
	mockES := new(MockElasticsearchClient)
	config := &Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	service := NewInventoryService(mockRepo, mockES, config)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)
	handler := NewHandler(service, catalog, nil)

	router := gin.New()
	router.GET("/api/inventory", handler.GetInventoryHandler)

	part1 := Part{
		PartCore: PartCore{
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

	// Настраиваем ожидание Elasticsearch:
	// При поиске "АКПП ACV30" сервис должен обратиться к mockES.SearchParts
	mockES.On("SearchParts", mock.MatchedBy(func(query map[string]interface{}) bool {
		boolQ, ok := query["bool"].(map[string]interface{})
		if !ok {
			return false
		}
		must, ok := boolQ["must"].([]map[string]interface{})
		if !ok || len(must) != 2 { // 2 слова: "АКПП" и "ACV30"
			return false
		}
		return true
	}), 0, 20).Return([]ElasticsearchPart{
		{ID: part1.ID, Name: part1.Name},
	}, int64(1), nil)

	// Сервис затем подгружает найденные ID из репозитория
	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		ids, ok := filters["ids_in"].([]int64)
		return ok && len(ids) == 1 && ids[0] == 10
	}), 0, 0).Return([]Part{part1}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?search=АКПП+ACV30&page=1&limit=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, int64(10), response[0].ID)
	assert.Equal(t, "АКПП", response[0].Name)
	assert.Equal(t, "Camry ACV30", response[0].Model)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestSearchIntegration_ElasticsearchFallbackToDatabase тестирует бесшовный откат на базу данных при ошибке Elasticsearch
func TestSearchIntegration_ElasticsearchFallbackToDatabase(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockPartRepository)
	mockES := new(MockElasticsearchClient)
	config := &Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	service := NewInventoryService(mockRepo, mockES, config)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)
	handler := NewHandler(service, catalog, nil)

	router := gin.New()
	router.GET("/api/inventory", handler.GetInventoryHandler)

	fallbackPart := Part{
		PartCore: PartCore{
			ID:       20,
			Name:     "Бампер передний",
			Quantity: 1,
			Price:    15000.0,
			Category: "Кузов снаружи",
			Brand:    "Honda",
			Model:    "Civic",
		},
	}

	// Elasticsearch возвращает ошибку (например, падение ноды или таймаут)
	mockES.On("SearchParts", mock.Anything, 0, 10).Return(
		[]ElasticsearchPart{},
		int64(0),
		errors.New("elasticsearch connection refused"),
	)

	// Сервис должен выполнить fallback в базу данных с параметрами поиска
	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		searchVal, ok := filters["search"].(string)
		return ok && searchVal == "Бампер"
	}), 0, 10).Return([]Part{fallbackPart}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?search=Бампер&page=1&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, int64(20), response[0].ID)
	assert.Equal(t, "Бампер передний", response[0].Name)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}

// TestSearchIntegration_CategoryFilter тестирует фильтрацию по категории через HTTP
func TestSearchIntegration_CategoryFilter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := new(MockPartRepository)
	mockES := new(MockElasticsearchClient)
	config := &Config{RedisURL: "127.0.0.1:6379"}

	mockRepo.On("GetTotalEarnings", mock.Anything).Return(1000.0, nil)
	mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)

	service := NewInventoryService(mockRepo, mockES, config)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)
	handler := NewHandler(service, catalog, nil)

	router := gin.New()
	router.GET("/api/inventory", handler.GetInventoryHandler)

	part := Part{
		PartCore: PartCore{
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
		// Проверяем наличие фильтра по категории
		catFilter := filters[1]["bool"].(map[string]interface{})
		shoulds := catFilter["should"].([]map[string]interface{})
		termClause := shoulds[0]["term"].(map[string]interface{})
		return termClause["category"] == "Тормозная система"
	}), 0, 50).Return([]ElasticsearchPart{
		{ID: part.ID, Name: part.Name},
	}, int64(1), nil)

	mockRepo.On("FindWithFilters", mock.Anything, mock.MatchedBy(func(filters map[string]interface{}) bool {
		ids, ok := filters["ids_in"].([]int64)
		return ok && len(ids) == 1 && ids[0] == 30
	}), 0, 0).Return([]Part{part}, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?category=Тормозная+система", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []Part
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	require.Len(t, response, 1)
	assert.Equal(t, "Тормозная система", response[0].Category)

	mockES.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
