package main

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockElasticsearchClient - мок для Elasticsearch клиента
type MockElasticsearchClient struct {
	mock.Mock
}

func (m *MockElasticsearchClient) IndexPart(part *Part) error {
	args := m.Called(part)
	return args.Error(0)
}

func (m *MockElasticsearchClient) DeletePartFromIndex(partID int64) error {
	args := m.Called(partID)
	return args.Error(0)
}

func (m *MockElasticsearchClient) SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	args := m.Called(query, from, size)
	return args.Get(0).([]ElasticsearchPart), args.Get(1).(int64), args.Error(2)
}

// ServiceTestSuite - набор тестов для сервиса
type ServiceTestSuite struct {
	suite.Suite
	mockRepo *MockPartRepository
	mockES   *MockElasticsearchClient
	service  InventoryService
	testPart *Part
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockPartRepository)
	suite.mockES = new(MockElasticsearchClient)
	config := &Config{RedisURL: "127.0.0.1:6379"}

	// Настраиваем mock для GetTotalEarnings, который вызывается в NewInventoryService
	suite.mockRepo.On("GetTotalEarnings", mock.Anything).Return(500.0, nil)

	suite.service = NewInventoryService(suite.mockRepo, suite.mockES, config)

	// Создаем тестовую запчасть
	suite.testPart = &Part{
		PartCore: PartCore{
			ID:          1,
			Name:        "Test Part",
			Quantity:    10,
			Description: "Test description",
			Category:    "Test Category",
			Price:       100.0,
		},
	}
}

func (suite *ServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockES.AssertExpectations(suite.T())
}

// TestAddPart - тест добавления запчасти
func (suite *ServiceTestSuite) TestAddPart() {
	newPart := &Part{
		PartCore: PartCore{
			Name:     "New Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	// Настраиваем моки
	suite.mockRepo.On("Create", mock.Anything, newPart).Return(nil).Run(func(args mock.Arguments) {
		part := args.Get(1).(*Part)
		part.ID = 2
	})
	suite.mockRepo.On("FindByID", mock.Anything, int64(2)).Return(newPart, nil)

	result, err := suite.service.AddPart(context.Background(), newPart)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New Part", result.Name)
}

// TestAddPart_InvalidData - тест добавления с невалидными данными
func (suite *ServiceTestSuite) TestAddPart_InvalidData() {
	invalidPart := &Part{
		PartCore: PartCore{
			Name:     "", // пустое имя - невалидно
			Quantity: 5,
		},
	}

	suite.mockRepo.On("Create", mock.Anything, invalidPart).Return(assert.AnError)

	result, err := suite.service.AddPart(context.Background(), invalidPart)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestUpdatePart - тест обновления запчасти
func (suite *ServiceTestSuite) TestUpdatePart() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	// Настраиваем моки
	suite.mockRepo.On("Update", mock.Anything, suite.testPart.ID, updates).Return(nil)
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	err := suite.service.UpdatePart(context.Background(), suite.testPart.ID, updates)
	assert.NoError(suite.T(), err)
}

// TestDeletePart - тест удаления запчасти
func (suite *ServiceTestSuite) TestDeletePart() {
	// Настраиваем моки
	suite.mockRepo.On("Delete", mock.Anything, suite.testPart.ID).Return(nil)
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	err := suite.service.DeletePart(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
}

// TestMarkPartForDeletion - тест отметки для удаления
func (suite *ServiceTestSuite) TestMarkPartForDeletion() {
	suite.mockRepo.On("MarkForDeletion", mock.Anything, suite.testPart.ID, mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.service.MarkPartForDeletion(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
}

// TestGetInventory - тест получения инвентаря
func (suite *ServiceTestSuite) TestGetInventory() {
	params := InventoryQueryParams{
		Page:  1,
		Limit: 10,
	}

	expectedParts := []Part{*suite.testPart}
	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)
	suite.mockRepo.On("FindWithFilters", mock.Anything, mock.Anything, 0, 10).Return(expectedParts, nil)

	parts, err := suite.service.GetInventory(context.Background(), params)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), parts)
	assert.Equal(suite.T(), suite.testPart.Name, parts[0].Name)
}

// TestGetInventory_WithSearch - тест поиска в инвентаре
func (suite *ServiceTestSuite) TestGetInventory_WithSearch() {
	params := InventoryQueryParams{
		Search: "Test",
		Page:   1,
		Limit:  10,
	}

	expectedParts := []Part{*suite.testPart}
	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)
	suite.mockES.On("SearchParts", mock.AnythingOfType("map[string]interface {}"), 0, 10).Return([]ElasticsearchPart{{ID: suite.testPart.ID, Name: suite.testPart.Name}}, int64(1), nil)
	suite.mockRepo.On("FindWithFilters", mock.Anything, mock.Anything, 0, 0).Return(expectedParts, nil)

	parts, err := suite.service.GetInventory(context.Background(), params)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), parts, 1)
	assert.Equal(suite.T(), suite.testPart.Name, parts[0].Name)
}

// TestGetStatistics - тест получения статистики
func (suite *ServiceTestSuite) TestGetStatistics() {
	suite.mockRepo.On("GetStatistics", mock.Anything).Return(StatisticsResponse{}, nil)
	suite.mockRepo.On("GetTotalEarnings", mock.Anything).Return(500.0, nil)

	stats, err := suite.service.GetStatistics(context.Background())
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.GreaterOrEqual(suite.T(), stats.TotalEarnings, 500.0)
}

// TestBulkDeleteParts - тест массового удаления
func (suite *ServiceTestSuite) TestBulkDeleteParts() {
	part2 := &Part{PartCore: PartCore{ID: 2, Name: "Part 2", Quantity: 5}}
	ids := []int64{suite.testPart.ID, part2.ID}

	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)
	suite.mockRepo.On("FindByID", mock.Anything, part2.ID).Return(part2, nil)
	suite.mockRepo.On("BulkDelete", mock.Anything, ids).Return(nil)

	err := suite.service.BulkDeleteParts(context.Background(), ids)
	assert.NoError(suite.T(), err)
}

// TestBulkUpdateParts - тест массового обновления
func (suite *ServiceTestSuite) TestBulkUpdateParts() {
	updates := []map[string]interface{}{
		{
			"id":       suite.testPart.ID,
			"name":     "Bulk Updated",
			"quantity": 99,
		},
	}

	suite.mockRepo.On("BulkUpdate", mock.Anything, updates).Return(1, nil)

	count, err := suite.service.BulkUpdateParts(context.Background(), updates)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

// TestGetPartByID - тест получения запчасти по ID
func (suite *ServiceTestSuite) TestGetPartByID() {
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	part, err := suite.service.GetPartByID(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), part)
	assert.Equal(suite.T(), suite.testPart.Name, part.Name)
}

// TestUpdateEarnings - тест обновления заработка
func (suite *ServiceTestSuite) TestUpdateEarnings() {
	suite.mockRepo.On("UpdateTotalEarnings", mock.Anything, mock.Anything).Return(nil)

	err := suite.service.UpdateEarnings(context.Background(), 100.0)
	assert.NoError(suite.T(), err)
}

// TestBuildElasticsearchQuery_MultiTerm - тест генерации запроса с раздельными словами
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_MultiTerm() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Search: "АКПП ACV30",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery, ok := query["bool"].(map[string]interface{})
	assert.True(suite.T(), ok)

	mustClauses, ok := boolQuery["must"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	// Должно быть 2 must clauses: по одному для каждого терма ("АКПП" и "ACV30")
	assert.Equal(suite.T(), 2, len(mustClauses))

	// Должны присутствовать should-клаузы верхнего уровня для релевантности
	shouldClauses, ok := boolQuery["should"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	assert.NotEmpty(suite.T(), shouldClauses)
}

// TestBuildElasticsearchQuery_Category - тест фильтрации по категории
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_Category() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Category: "Тормозная система",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery, ok := query["bool"].(map[string]interface{})
	assert.True(suite.T(), ok)

	filters, ok := boolQuery["filter"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	assert.GreaterOrEqual(suite.T(), len(filters), 2)
}

// TestRunSuite - запуск всех тестов сервиса
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
