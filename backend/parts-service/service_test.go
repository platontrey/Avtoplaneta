package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// MockElasticsearchClient - мок для Elasticsearch клиента
type MockElasticsearchClient struct {
	mock.Mock
}

func (m *MockElasticsearchClient) IndexPart(part *Part) error {
	args := m.Called(part)
	return args.Error(0)
}

func (m *MockElasticsearchClient) DeletePartFromIndex(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockElasticsearchClient) SearchParts(query map[string]interface{}, from, size int) ([]Part, int, error) {
	args := m.Called(query, from, size)
	return args.Get(0).([]Part), args.Int(1), args.Error(2)
}

// ServiceTestSuite - набор тестов для сервиса
type ServiceTestSuite struct {
	suite.Suite
	db         *gorm.DB
	repo       PartRepository
	mockES     *MockElasticsearchClient
	service    InventoryService
	testPart   *Part
}

func (suite *ServiceTestSuite) SetupTest() {
	// Создаем in-memory базу данных
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Миграция схемы
	err = db.AutoMigrate(&Part{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewPartRepository(db)
	suite.mockES = new(MockElasticsearchClient)
	suite.service = NewInventoryService(suite.repo, suite.mockES)

	// Создаем тестовую запчасть
	suite.testPart = &Part{
		Name:        "Test Part",
		Quantity:    10,
		Description: "Test description",
		Category:    "Test Category",
		Price:       100.0,
	}
	err = suite.repo.Create(suite.testPart)
	suite.Require().NoError(err)
}

func (suite *ServiceTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// TestAddPart - тест добавления запчасти
func (suite *ServiceTestSuite) TestAddPart() {
	newPart := &Part{
		Name:     "New Part",
		Quantity: 5,
		Price:    50.0,
	}

	// Настраиваем мок для Elasticsearch
	suite.mockES.On("IndexPart", mock.AnythingOfType("*main.Part")).Return(nil)

	result, err := suite.service.AddPart(newPart)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New Part", result.Name)

	suite.mockES.AssertExpectations(suite.T())
}

// TestAddPart_InvalidData - тест добавления с невалидными данными
func (suite *ServiceTestSuite) TestAddPart_InvalidData() {
	invalidPart := &Part{
		Name:     "", // пустое имя - невалидно
		Quantity: 5,
	}

	result, err := suite.service.AddPart(invalidPart)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestUpdatePart - тест обновления запчасти
func (suite *ServiceTestSuite) TestUpdatePart() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	// Настраиваем мок для Elasticsearch
	suite.mockES.On("IndexPart", mock.AnythingOfType("*main.Part")).Return(nil)

	err := suite.service.UpdatePart(suite.testPart.ID, updates)
	assert.NoError(suite.T(), err)

	// Проверяем обновление
	updated, err := suite.repo.FindByID(suite.testPart.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", updated.Name)
	assert.Equal(suite.T(), 20, updated.Quantity)

	suite.mockES.AssertExpectations(suite.T())
}

// TestDeletePart - тест удаления запчасти
func (suite *ServiceTestSuite) TestDeletePart() {
	// Настраиваем мок для Elasticsearch
	suite.mockES.On("DeletePartFromIndex", suite.testPart.ID).Return(nil)

	err := suite.service.DeletePart(suite.testPart.ID)
	assert.NoError(suite.T(), err)

	// Проверяем, что удалена
	_, err = suite.repo.FindByID(suite.testPart.ID)
	assert.Error(suite.T(), err)

	suite.mockES.AssertExpectations(suite.T())
}

// TestMarkPartForDeletion - тест отметки для удаления
func (suite *ServiceTestSuite) TestMarkPartForDeletion() {
	err := suite.service.MarkPartForDeletion(suite.testPart.ID)
	assert.NoError(suite.T(), err)

	// Проверяем
	updated, err := suite.repo.FindByID(suite.testPart.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated.ToDeleteAt)
}

// TestGetInventory - тест получения инвентаря
func (suite *ServiceTestSuite) TestGetInventory() {
	params := InventoryQueryParams{
		Page:  1,
		Limit: 10,
	}

	parts, err := suite.service.GetInventory(params)
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

	parts, err := suite.service.GetInventory(params)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), parts, 1)
	assert.Equal(suite.T(), suite.testPart.Name, parts[0].Name)
}

// TestGetStatistics - тест получения статистики
func (suite *ServiceTestSuite) TestGetStatistics() {
	stats, err := suite.service.GetStatistics()
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.GreaterOrEqual(suite.T(), stats.TotalParts, int64(1))
}

// TestBulkDeleteParts - тест массового удаления
func (suite *ServiceTestSuite) TestBulkDeleteParts() {
	// Создаем еще одну запчасть
	part2 := &Part{Name: "Part 2", Quantity: 5}
	err := suite.repo.Create(part2)
	suite.Require().NoError(err)

	ids := []uint{suite.testPart.ID, part2.ID}
	err = suite.service.BulkDeleteParts(ids)
	assert.NoError(suite.T(), err)

	// Проверяем, что обе удалены
	_, err1 := suite.repo.FindByID(suite.testPart.ID)
	_, err2 := suite.repo.FindByID(part2.ID)
	assert.Error(suite.T(), err1)
	assert.Error(suite.T(), err2)
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

	err := suite.service.BulkUpdateParts(updates)
	assert.NoError(suite.T(), err)

	// Проверяем обновление
	updated, err := suite.repo.FindByID(suite.testPart.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Bulk Updated", updated.Name)
	assert.Equal(suite.T(), 99, updated.Quantity)
}

// TestRunSuite - запуск всех тестов сервиса
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}