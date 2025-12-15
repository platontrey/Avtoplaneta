package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockInventoryService - мок для InventoryService
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) GetInventory(params InventoryQueryParams) ([]Part, error) {
	args := m.Called(params)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockInventoryService) AddPart(part *Part) (*Part, error) {
	args := m.Called(part)
	return args.Get(0).(*Part), args.Error(1)
}

func (m *MockInventoryService) UpdatePart(id uint, updates map[string]interface{}) error {
	args := m.Called(id, updates)
	return args.Error(0)
}

func (m *MockInventoryService) DeletePart(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryService) MarkPartForDeletion(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockInventoryService) GetStatistics() (StatisticsResponse, error) {
	args := m.Called()
	return args.Get(0).(StatisticsResponse), args.Error(1)
}

func (m *MockInventoryService) BulkDeleteParts(ids []uint) error {
	args := m.Called(ids)
	return args.Error(0)
}

func (m *MockInventoryService) BulkUpdateParts(updates []map[string]interface{}) error {
	args := m.Called(updates)
	return args.Error(0)
}

func (m *MockInventoryService) DeleteZeroQuantityPartsBySupplier(supplierCode string) (int64, error) {
	args := m.Called(supplierCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockInventoryService) GetSupplierCodes() ([]string, error) {
	args := m.Called()
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockInventoryService) UploadPartPhoto(id uint, c *gin.Context) (string, error) {
	args := m.Called(id, c)
	return args.String(0), args.Error(1)
}

func (m *MockInventoryService) DeletePartPhoto(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// HandlersTestSuite - набор тестов для handlers
type HandlersTestSuite struct {
	suite.Suite
	mockService *MockInventoryService
	handler     *Handler
	router      *gin.Engine
}

func (suite *HandlersTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(MockInventoryService)
	suite.handler = NewHandler(suite.mockService)
	suite.router = gin.New()

	// Настраиваем маршруты
	suite.router.GET("/api/inventory", suite.handler.GetInventoryHandler)
	suite.router.POST("/api/addpart", suite.handler.AddPartHandler)
	suite.router.PUT("/api/updatepart/:id", suite.handler.UpdatePartHandler)
	suite.router.DELETE("/api/deletepart/:id", suite.handler.DeletePartHandler)
}

func (suite *HandlersTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// TestGetInventoryHandler - тест получения инвентаря
func (suite *HandlersTestSuite) TestGetInventoryHandler() {
	expectedParts := []Part{
		{PartCore: PartCore{ID: 1, Name: "Test Part", Quantity: 10}},
	}

	suite.mockService.On("GetInventory", mock.AnythingOfType("main.InventoryQueryParams")).Return(expectedParts, nil)

	req, _ := http.NewRequest("GET", "/api/inventory?page=1&limit=10", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response []Part
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response, 1)
	assert.Equal(suite.T(), "Test Part", response[0].Name)
}

// TestAddPartHandler - тест добавления запчасти
func (suite *HandlersTestSuite) TestAddPartHandler() {
	newPart := &Part{
		PartCore: PartCore{
			Name:     "New Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	expectedResult := &Part{
		PartCore: PartCore{
			ID:       1,
			Name:     "New Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	suite.mockService.On("AddPart", mock.AnythingOfType("*main.Part")).Return(expectedResult, nil)

	partJSON, _ := json.Marshal(newPart)
	req, _ := http.NewRequest("POST", "/api/addpart", bytes.NewBuffer(partJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Часть добавлена успешно", response["message"])
	assert.Equal(suite.T(), float64(1), response["id"])
}

// TestAddPartHandler_InvalidJSON - тест с невалидным JSON
func (suite *HandlersTestSuite) TestAddPartHandler_InvalidJSON() {
	req, _ := http.NewRequest("POST", "/api/addpart", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestUpdatePartHandler - тест обновления запчасти
func (suite *HandlersTestSuite) TestUpdatePartHandler() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	suite.mockService.On("UpdatePart", uint(1), updates).Return(nil)

	updatesJSON, _ := json.Marshal(updates)
	req, _ := http.NewRequest("PUT", "/api/updatepart/1", bytes.NewBuffer(updatesJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Часть обновлена успешно", response["message"])
}

// TestUpdatePartHandler_InvalidID - тест с невалидным ID
func (suite *HandlersTestSuite) TestUpdatePartHandler_InvalidID() {
	updates := map[string]interface{}{"name": "Test"}
	updatesJSON, _ := json.Marshal(updates)

	req, _ := http.NewRequest("PUT", "/api/updatepart/invalid", bytes.NewBuffer(updatesJSON))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// TestDeletePartHandler - тест удаления запчасти
func (suite *HandlersTestSuite) TestDeletePartHandler() {
	suite.mockService.On("DeletePart", uint(1)).Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/deletepart/1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Запчасть удалена успешно", response["message"])
}

// TestGetStatisticsHandler - тест получения статистики
func (suite *HandlersTestSuite) TestGetStatisticsHandler() {
	expectedStats := StatisticsResponse{
		TotalParts:    10,
		TotalValue:    1000.0,
		TotalEarnings: 500.0,
		Categories: []CategoryCount{
			{Name: "Category 1", Count: 5},
		},
	}

	suite.mockService.On("GetStatistics").Return(expectedStats, nil)

	req, _ := http.NewRequest("GET", "/api/statistics", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response StatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(10), response.TotalParts)
	assert.Equal(suite.T(), 1000.0, response.TotalValue)
}

// TestBulkDeletePartsHandler - тест массового удаления
func (suite *HandlersTestSuite) TestBulkDeletePartsHandler() {
	ids := []uint{1, 2, 3}

	suite.mockService.On("BulkDeleteParts", ids).Return(nil)

	requestData := map[string]interface{}{"ids": ids}
	jsonData, _ := json.Marshal(requestData)

	req, _ := http.NewRequest("DELETE", "/api/admin/bulk-delete-parts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Запчасти удалены успешно", response["message"])
}

// TestRunSuite - запуск всех тестов handlers
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}