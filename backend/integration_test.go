package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// IntegrationTestSuite - набор интеграционных тестов
type IntegrationTestSuite struct {
	suite.Suite
	db         *gorm.DB
	router     *gin.Engine
	testPartID uint
}

func (suite *IntegrationTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Создаем in-memory базу данных
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Миграция схемы
	err = db.AutoMigrate(&Part{})
	suite.Require().NoError(err)

	suite.db = db

	// Инициализируем сервисы
	repo := NewPartRepository(db)
	es := &MockElasticsearchClient{} // Используем мок для ES в интеграционных тестах
	service := NewInventoryService(repo, es)
	handler := NewHandler(service)

	// Настраиваем маршруты
	suite.router = gin.New()
	suite.router.GET("/api/inventory", handler.GetInventoryHandler)
	suite.router.POST("/api/addpart", handler.AddPartHandler)
	suite.router.PUT("/api/updatepart/:id", handler.UpdatePartHandler)
	suite.router.DELETE("/api/deletepart/:id", handler.DeletePartHandler)
	suite.router.POST("/api/markpartfordeletion/:id", handler.MarkPartForDeletionHandler)
	suite.router.GET("/api/statistics", handler.GetStatisticsHandler)
	suite.router.DELETE("/api/admin/bulk-delete-parts", handler.BulkDeletePartsHandler)
	suite.router.PUT("/api/admin/bulk-update-parts", handler.BulkUpdatePartsHandler)

	// Создаем тестовую запчасть
	testPart := &Part{
		Name:        "Integration Test Part",
		Quantity:    10,
		Description: "Test part for integration tests",
		Category:    "Test Category",
		Price:       100.0,
	}
	err = db.Create(testPart).Error
	suite.Require().NoError(err)
	suite.testPartID = testPart.ID
}

func (suite *IntegrationTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// TestFullCRUDCycle - тест полного цикла CRUD операций
func (suite *IntegrationTestSuite) TestFullCRUDCycle() {
	// 1. Создание новой запчасти
	newPart := &Part{
		Name:     "CRUD Test Part",
		Quantity: 5,
		Price:    50.0,
	}
	partJSON, _ := json.Marshal(newPart)

	req, _ := http.NewRequest("POST", "/api/addpart", bytes.NewBuffer(partJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var createResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &createResponse)
	createdID := uint(createResponse["id"].(float64))

	// 2. Получение инвентаря (должна содержать обе запчасти)
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var inventory []Part
	json.Unmarshal(w.Body.Bytes(), &inventory)
	assert.GreaterOrEqual(suite.T(), len(inventory), 2)

	// 3. Обновление запчасти
	updates := map[string]interface{}{
		"name":     "Updated CRUD Part",
		"quantity": 15,
	}
	updateJSON, _ := json.Marshal(updates)

	req, _ = http.NewRequest("PUT", "/api/updatepart/"+string(rune(createdID)), bytes.NewBuffer(updateJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 4. Проверка обновления через получение инвентаря
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &inventory)
	found := false
	for _, part := range inventory {
		if part.ID == createdID {
			assert.Equal(suite.T(), "Updated CRUD Part", part.Name)
			assert.Equal(suite.T(), 15, part.Quantity)
			found = true
			break
		}
	}
	assert.True(suite.T(), found, "Updated part not found in inventory")

	// 5. Удаление запчасти
	req, _ = http.NewRequest("DELETE", "/api/deletepart/"+string(rune(createdID)), nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// 6. Проверка удаления
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	json.Unmarshal(w.Body.Bytes(), &inventory)
	for _, part := range inventory {
		assert.NotEqual(suite.T(), createdID, part.ID, "Deleted part still exists")
	}
}

// TestStatisticsIntegration - тест интеграции статистики
func (suite *IntegrationTestSuite) TestStatisticsIntegration() {
	// Создаем дополнительные запчасти для статистики
	parts := []*Part{
		{Name: "Stat Part 1", Quantity: 5, Price: 25.0, Category: "Electronics"},
		{Name: "Stat Part 2", Quantity: 3, Price: 75.0, Category: "Electronics"},
		{Name: "Stat Part 3", Quantity: 2, Price: 50.0, Category: "Tools"},
	}

	for _, part := range parts {
		suite.db.Create(part)
	}

	req, _ := http.NewRequest("GET", "/api/statistics", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var stats StatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(suite.T(), err)

	// Проверяем, что статистика корректна
	assert.GreaterOrEqual(suite.T(), stats.TotalParts, int64(4)) // исходная + 3 новые
	assert.Greater(suite.T(), stats.TotalValue, 0.0)
	assert.GreaterOrEqual(suite.T(), len(stats.Categories), 2) // минимум 2 категории
}

// TestBulkOperationsIntegration - тест массовых операций
func (suite *IntegrationTestSuite) TestBulkOperationsIntegration() {
	// Создаем дополнительные запчасти для массовых операций
	bulkParts := []*Part{
		{Name: "Bulk Part 1", Quantity: 1},
		{Name: "Bulk Part 2", Quantity: 1},
		{Name: "Bulk Part 3", Quantity: 1},
	}

	var bulkIDs []uint
	for _, part := range bulkParts {
		suite.db.Create(part)
		bulkIDs = append(bulkIDs, part.ID)
	}

	// Массовое обновление
	bulkUpdates := []map[string]interface{}{
		{"id": bulkIDs[0], "name": "Bulk Updated 1", "quantity": 10},
		{"id": bulkIDs[1], "name": "Bulk Updated 2", "quantity": 20},
	}
	updatesJSON, _ := json.Marshal(bulkUpdates)

	req, _ := http.NewRequest("PUT", "/api/admin/bulk-update-parts", bytes.NewBuffer(updatesJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Проверяем обновления
	for i, id := range bulkIDs[:2] {
		var part Part
		suite.db.First(&part, id)
		assert.Equal(suite.T(), "Bulk Updated "+string(rune(i+1)), part.Name)
		assert.Equal(suite.T(), (i+1)*10, part.Quantity)
	}

	// Массовое удаление
	bulkDeleteData := map[string]interface{}{"ids": bulkIDs}
	deleteJSON, _ := json.Marshal(bulkDeleteData)

	req, _ = http.NewRequest("DELETE", "/api/admin/bulk-delete-parts", bytes.NewBuffer(deleteJSON))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Проверяем удаление
	for _, id := range bulkIDs {
		var count int64
		suite.db.Model(&Part{}).Where("id = ?", id).Count(&count)
		assert.Equal(suite.T(), int64(0), count)
	}
}

// TestMarkForDeletionIntegration - тест интеграции отметки для удаления
func (suite *IntegrationTestSuite) TestMarkForDeletionIntegration() {
	req, _ := http.NewRequest("POST", "/api/markpartfordeletion/"+string(rune(suite.testPartID)), nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Проверяем, что запчасть отмечена для удаления
	var part Part
	suite.db.First(&part, suite.testPartID)
	assert.NotNil(suite.T(), part.ToDeleteAt)

	// Имитируем очистку (в реальном приложении это делает планировщик)
	// Устанавливаем дату удаления в прошлое
	pastTime := time.Now().AddDate(0, 0, -15)
	suite.db.Model(&part).Update("to_delete_at", pastTime)

	// Запускаем получение инвентаря (должно удалить просроченные)
	req, _ = http.NewRequest("GET", "/api/inventory", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Проверяем, что запчасть удалена
	var count int64
	suite.db.Model(&Part{}).Where("id = ?", suite.testPartID).Count(&count)
	assert.Equal(suite.T(), int64(0), count)
}

// TestErrorHandlingIntegration - тест обработки ошибок
func (suite *IntegrationTestSuite) TestErrorHandlingIntegration() {
	// Тест с невалидным JSON
	req, _ := http.NewRequest("POST", "/api/addpart", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	// Тест с несуществующим ID
	req, _ = http.NewRequest("DELETE", "/api/deletepart/99999", nil)
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusInternalServerError, w.Code)

	// Тест с невалидным ID в URL
	req, _ = http.NewRequest("PUT", "/api/updatepart/invalid", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestRunSuite - запуск интеграционных тестов
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}