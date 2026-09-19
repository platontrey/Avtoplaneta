package main

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
	"github.com/stretchr/testify/suite"
)

// MockInventoryService - мок для InventoryService
type MockInventoryService struct {
	mock.Mock
}

func (m *MockInventoryService) GetInventory(ctx context.Context, params InventoryQueryParams) ([]Part, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockInventoryService) AddPart(ctx context.Context, part *Part) (*Part, error) {
	args := m.Called(ctx, part)
	return args.Get(0).(*Part), args.Error(1)
}

func (m *MockInventoryService) UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockInventoryService) DeletePart(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInventoryService) MarkPartForDeletion(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInventoryService) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(StatisticsResponse), args.Error(1)
}

// Версия склада в юнит-тестах не участвует: пустая строка отключает условный
// ответ, и обработчик идёт по обычному пути.
func (m *MockInventoryService) InventoryVersion(context.Context) (string, error) {
	return "", nil
}

func (m *MockInventoryService) BulkDeleteParts(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockInventoryService) BulkUpdateParts(ctx context.Context, updates []map[string]interface{}) (int, error) {
	args := m.Called(ctx, updates)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockInventoryService) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	args := m.Called(ctx, supplierCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockInventoryService) GetSupplierCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockInventoryService) UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) {
	args := m.Called(ctx, id, c)
	return args.String(0), args.Error(1)
}

func (m *MockInventoryService) DeletePartPhoto(ctx context.Context, id int64, photoPath string) error {
	args := m.Called(ctx, id, photoPath)
	return args.Error(0)
}

func (m *MockInventoryService) GetPartByID(ctx context.Context, id int64) (*Part, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Part), args.Error(1)
}

func (m *MockInventoryService) AddPartsBatch(ctx context.Context, parts []Part) ([]Part, error) {
	args := m.Called(ctx, parts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockInventoryService) DecreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockInventoryService) IncreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockInventoryService) UpdateEarnings(ctx context.Context, amount float64) error {
	args := m.Called(ctx, amount)
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
	suite.handler = NewHandler(suite.mockService, nil, nil, nil)
	suite.router = gin.New()

	// Настраиваем маршруты
	suite.router.GET("/api/inventory", suite.handler.GetInventoryHandler)
	suite.router.POST("/api/addpart", suite.handler.AddPartHandler)
	suite.router.DELETE("/api/deletepart/:id", suite.handler.DeletePartHandler)
	suite.router.PUT("/api/updatepart/:id", suite.handler.UpdatePartHandler)
	suite.router.POST("/api/uploadpartphoto/:id", suite.handler.UploadPartPhotoHandler)
	suite.router.DELETE("/api/deletepartphoto/:id", suite.handler.DeletePartPhotoHandler)
	suite.router.POST("/api/markpartfordeletion/:id", suite.handler.MarkPartForDeletionHandler)
	suite.router.GET("/api/statistics", suite.handler.GetStatisticsHandler)
	suite.router.POST("/api/statistics/update-earnings", suite.handler.UpdateEarningsHandler)
	suite.router.DELETE("/api/admin/delete-zero-quantity-parts/:supplier_code", suite.handler.DeleteZeroQuantityPartsBySupplierHandler)
	suite.router.GET("/api/admin/supplier-codes", suite.handler.GetSupplierCodesHandler)
	suite.router.DELETE("/api/admin/bulk-delete-parts", suite.handler.BulkDeletePartsHandler)
	suite.router.PUT("/api/admin/bulk-update-parts", suite.handler.BulkUpdatePartsHandler)
}

func (suite *HandlersTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
}

// TestGetInventoryHandler - тест получения инвентаря
func (suite *HandlersTestSuite) TestGetInventoryHandler() {
	expectedParts := []Part{
		{PartCore: PartCore{ID: 1, Name: "Test Part", Quantity: 10}},
	}

	suite.mockService.On("GetInventory", mock.Anything, mock.AnythingOfType("main.InventoryQueryParams")).Return(expectedParts, nil)

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

	suite.mockService.On("AddPart", mock.Anything, mock.AnythingOfType("*main.Part")).Return(expectedResult, nil)

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

	suite.mockService.On("UpdatePart", mock.Anything, int64(1), map[string]interface{}{"name": "Updated Name", "quantity": float64(20)}).Return(nil)

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
	suite.mockService.On("DeletePart", mock.Anything, int64(1)).Return(nil)
	suite.mockService.On("GetPartByID", mock.Anything, int64(1)).Return(&Part{PartCore: PartCore{Name: "Test Part"}}, nil)

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
		TotalQuantity: 50,
		TotalValue:    1000.0,
		TotalEarnings: 500.0,
		Categories: []CategoryCount{
			{Name: "Category 1", Count: 5},
		},
	}

	suite.mockService.On("GetStatistics", mock.Anything).Return(expectedStats, nil)

	req, _ := http.NewRequest("GET", "/api/statistics", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response StatisticsResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 10, response.TotalParts)
	assert.Equal(suite.T(), 1000.0, response.TotalValue)
}

// TestBulkDeletePartsHandler - тест массового удаления
func (suite *HandlersTestSuite) TestBulkDeletePartsHandler() {
	ids := []int64{1, 2, 3}

	suite.mockService.On("BulkDeleteParts", mock.Anything, ids).Return(nil)

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

// TestMarkPartForDeletionHandler - тест отметки для удаления
func (suite *HandlersTestSuite) TestMarkPartForDeletionHandler() {
	suite.mockService.On("MarkPartForDeletion", mock.Anything, int64(1)).Return(nil)

	req, _ := http.NewRequest("POST", "/api/markpartfordeletion/1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Часть отмечена для удаления через 14 дней", response["message"])
}

// TestUploadPartPhotoHandler - тест загрузки фото
func (suite *HandlersTestSuite) TestUploadPartPhotoHandler() {
	suite.mockService.On("UploadPartPhoto", mock.Anything, int64(1), mock.Anything).Return("uploads/photo.jpg", nil)
	suite.mockService.On("GetPartByID", mock.Anything, int64(1)).Return(&Part{PartCore: PartCore{Photos: []string{"uploads/photo.jpg"}}}, nil)

	req, _ := http.NewRequest("POST", "/api/uploadpartphoto/1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Фото загружено успешно", response["message"])
	assert.Equal(suite.T(), "uploads/photo.jpg", response["photo"])
}

// TestDeletePartPhotoHandler - тест удаления фото
func (suite *HandlersTestSuite) TestDeletePartPhotoHandler() {
	suite.mockService.On("DeletePartPhoto", mock.Anything, int64(1), "").Return(nil)

	req, _ := http.NewRequest("DELETE", "/api/deletepartphoto/1", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Фото удалено успешно", response["message"])
}

// TestBulkUpdatePartsHandler - тест массового обновления
func (suite *HandlersTestSuite) TestBulkUpdatePartsHandler() {
	updates := []map[string]interface{}{
		{"id": 1, "name": "Updated Part"},
	}

	suite.mockService.On("BulkUpdateParts", mock.Anything, []map[string]interface{}{{"id": float64(1), "name": "Updated Part"}}).Return(1, nil)

	jsonData, _ := json.Marshal(updates)

	req, _ := http.NewRequest("PUT", "/api/admin/bulk-update-parts", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Запчасти обновлены успешно", response["message"])
	assert.Equal(suite.T(), float64(1), response["updated_count"])
}

// TestDeleteZeroQuantityPartsBySupplierHandler - тест удаления по поставщику
func (suite *HandlersTestSuite) TestDeleteZeroQuantityPartsBySupplierHandler() {
	suite.mockService.On("DeleteZeroQuantityPartsBySupplier", mock.Anything, "SUP001").Return(int64(5), nil)

	req, _ := http.NewRequest("DELETE", "/api/admin/delete-zero-quantity-parts/SUP001", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Запчасти с нулевым количеством удалены", response["message"])
	assert.Equal(suite.T(), float64(5), response["deleted_count"])
}

// TestGetSupplierCodesHandler - тест получения кодов поставщиков
func (suite *HandlersTestSuite) TestGetSupplierCodesHandler() {
	expectedCodes := []string{"SUP001", "SUP002"}

	suite.mockService.On("GetSupplierCodes", mock.Anything).Return(expectedCodes, nil)

	req, _ := http.NewRequest("GET", "/api/admin/supplier-codes", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	codesInterface := response["supplier_codes"].([]interface{})
	var codes []string
	for _, code := range codesInterface {
		codes = append(codes, code.(string))
	}
	assert.Equal(suite.T(), expectedCodes, codes)
}

// TestUpdateEarningsHandler - тест обновления заработка
func (suite *HandlersTestSuite) TestUpdateEarningsHandler() {
	reqData := map[string]float64{"amount": 100.0}
	jsonData, _ := json.Marshal(reqData)

	suite.mockService.On("UpdateEarnings", mock.Anything, 100.0).Return(nil)

	req, _ := http.NewRequest("POST", "/api/statistics/update-earnings", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Заработок обновлен", response["message"])
}

// TestRunSuite - запуск всех тестов handlers
func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

func (m *MockInventoryService) RenameSeller(context.Context, int64, string) ([]Part, error) {
	return nil, nil
}

func (m *MockInventoryService) PartsForExport(context.Context) ([]Part, error) {
	return nil, nil
}

func TestSetupRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	handler := &Handler{inventoryService: new(MockInventoryService)}
	assert.NotPanics(t, func() {
		SetupRoutes(r, handler)
	})

	// Test health endpoint
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)
}
