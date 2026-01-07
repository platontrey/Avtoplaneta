package main

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"gorm.io/gorm"
)

// MockPartRepository - мок для PartRepository
type MockPartRepository struct {
	mock.Mock
}

func (m *MockPartRepository) Create(ctx context.Context, part *Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockPartRepository) FindByID(ctx context.Context, id uint) (*Part, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*Part), args.Error(1)
}

func (m *MockPartRepository) FindAll(ctx context.Context, query *gorm.DB) ([]Part, error) {
	args := m.Called(ctx, query)
	if args.Get(0) == nil {
		return []Part{}, args.Error(1)
	}
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]Part, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockPartRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartRepository) MarkForDeletion(ctx context.Context, id uint, deleteAt time.Time) error {
	args := m.Called(ctx, id, deleteAt)
	return args.Error(0)
}

func (m *MockPartRepository) DeleteExpiredParts(ctx context.Context, before time.Time) error {
	args := m.Called(ctx, before)
	return args.Error(0)
}

func (m *MockPartRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockPartRepository) UpdateEarnings(ctx context.Context, amount float64) error {
	args := m.Called(ctx, amount)
	return args.Error(0)
}

func (m *MockPartRepository) BulkDelete(ctx context.Context, ids []uint) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockPartRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	args := m.Called(ctx, updates)
	return args.Get(0).(int), args.Error(1)
}

func (m *MockPartRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	args := m.Called(ctx, supplierCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartRepository) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(StatisticsResponse), args.Error(1)
}

func (m *MockPartRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	args := m.Called(ctx, amount)
	return args.Error(0)
}

func (m *MockPartRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

// RepositoryTestSuite - набор тестов для репозитория
type RepositoryTestSuite struct {
	suite.Suite
	mockRepo *MockPartRepository
	repo     PartRepository
}

// SetupTest - настройка перед каждым тестом
func (suite *RepositoryTestSuite) SetupTest() {
	suite.mockRepo = new(MockPartRepository)
	suite.repo = suite.mockRepo
}

// TearDownTest - очистка после каждого теста
func (suite *RepositoryTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

// TestCreate - тест создания запчасти
func (suite *RepositoryTestSuite) TestCreate() {
	part := &Part{
		PartCore: PartCore{
			Name:        "Test Part",
			Quantity:    10,
			Description: "Test description",
			Category:    "Test Category",
			Price:       100.0,
		},
	}

	suite.mockRepo.On("Create", mock.Anything, part).Return(nil)

	err := suite.repo.Create(context.Background(), part)
	assert.NoError(suite.T(), err)
}

// TestFindByID - тест поиска по ID
func (suite *RepositoryTestSuite) TestFindByID() {
	expectedPart := &Part{
		PartCore: PartCore{
			ID:       1,
			Name:     "Test Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	suite.mockRepo.On("FindByID", mock.Anything, uint(1)).Return(expectedPart, nil)

	found, err := suite.repo.FindByID(context.Background(), 1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), expectedPart.Name, found.Name)
	assert.Equal(suite.T(), expectedPart.Quantity, found.Quantity)
}

// TestFindByID_NotFound - тест поиска несуществующей запчасти
func (suite *RepositoryTestSuite) TestFindByID_NotFound() {
	suite.mockRepo.On("FindByID", mock.Anything, uint(999)).Return((*Part)(nil), assert.AnError)

	found, err := suite.repo.FindByID(context.Background(), 999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestUpdate - тест обновления запчасти
func (suite *RepositoryTestSuite) TestUpdate() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	suite.mockRepo.On("Update", mock.Anything, uint(1), updates).Return(nil)

	err := suite.repo.Update(context.Background(), 1, updates)
	assert.NoError(suite.T(), err)
}

// TestDelete - тест удаления запчасти
func (suite *RepositoryTestSuite) TestDelete() {
	suite.mockRepo.On("Delete", mock.Anything, uint(1)).Return(nil)

	err := suite.repo.Delete(context.Background(), 1)
	assert.NoError(suite.T(), err)
}

// TestMarkForDeletion - тест отметки для удаления
func (suite *RepositoryTestSuite) TestMarkForDeletion() {
	deleteAt := time.Now().AddDate(0, 0, 14)

	suite.mockRepo.On("MarkForDeletion", mock.Anything, uint(1), mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.repo.MarkForDeletion(context.Background(), 1, deleteAt)
	assert.NoError(suite.T(), err)
}

// TestDeleteExpiredParts - тест удаления просроченных запчастей
func (suite *RepositoryTestSuite) TestDeleteExpiredParts() {
	now := time.Now()

	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.repo.DeleteExpiredParts(context.Background(), now)
	assert.NoError(suite.T(), err)
}

// TestFindAll - тест получения всех запчастей
func (suite *RepositoryTestSuite) TestFindAll() {
	expectedParts := []Part{
		{PartCore: PartCore{Name: "Part 1", Quantity: 1}},
		{PartCore: PartCore{Name: "Part 2", Quantity: 2}},
		{PartCore: PartCore{Name: "Part 3", Quantity: 3}},
	}

	suite.mockRepo.On("FindAll", mock.Anything, mock.Anything).Return(expectedParts, nil)

	found, err := suite.repo.FindAll(context.Background(), nil)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 3)
}

// TestRunSuite - запуск всех тестов
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
