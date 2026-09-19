package main

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockPartRepository - мок для PartRepository
type MockPartRepository struct {
	mock.Mock
}

func (m *MockPartRepository) Create(ctx context.Context, part *Part) error {
	args := m.Called(ctx, part)
	return args.Error(0)
}

func (m *MockPartRepository) FindByID(ctx context.Context, id int64) (*Part, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Part), args.Error(1)
}

func (m *MockPartRepository) FindAll(ctx context.Context) ([]Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]Part, error) {
	args := m.Called(ctx, filters, offset, limit)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	args := m.Called(ctx, id, updates)
	return args.Error(0)
}

func (m *MockPartRepository) Delete(ctx context.Context, id int64) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockPartRepository) CreateBatch(ctx context.Context, parts []Part) ([]Part, error) {
	args := m.Called(ctx, parts)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockPartRepository) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	args := m.Called(ctx, id, amount, operationID)
	return args.Error(0)
}

func (m *MockPartRepository) MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error {
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

func (m *MockPartRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	args := m.Called(ctx, amount)
	return args.Error(0)
}

func (m *MockPartRepository) BulkDelete(ctx context.Context, ids []int64) error {
	args := m.Called(ctx, ids)
	return args.Error(0)
}

func (m *MockPartRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	args := m.Called(ctx, updates)
	return args.Int(0), args.Error(1)
}

func (m *MockPartRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	args := m.Called(ctx, supplierCode)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPartRepository) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	args := m.Called(ctx)
	return args.Get(0).(StatisticsResponse), args.Error(1)
}

func (m *MockPartRepository) InventoryVersion(context.Context) (string, error) {
	return "", nil
}

func (m *MockPartRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	args := m.Called(ctx)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockPartRepository) UpdatePartPhotos(ctx context.Context, id int64, photos StringArray) error {
	args := m.Called(ctx, id, photos)
	return args.Error(0)
}

func (m *MockPartRepository) GetPartsForXML(ctx context.Context) ([]Part, error) {
	args := m.Called(ctx)
	return args.Get(0).([]Part), args.Error(1)
}

func (m *MockPartRepository) GetLastCreatedPart(ctx context.Context) (*Part, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*Part), args.Error(1)
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

func (suite *RepositoryTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
}

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

func (suite *RepositoryTestSuite) TestFindByID() {
	expectedPart := &Part{
		PartCore: PartCore{
			ID:       1,
			Name:     "Test Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	suite.mockRepo.On("FindByID", mock.Anything, int64(1)).Return(expectedPart, nil)

	found, err := suite.repo.FindByID(context.Background(), 1)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), expectedPart.Name, found.Name)
	assert.Equal(suite.T(), expectedPart.Quantity, found.Quantity)
}

func (suite *RepositoryTestSuite) TestFindByID_NotFound() {
	suite.mockRepo.On("FindByID", mock.Anything, int64(999)).Return((*Part)(nil), assert.AnError)

	found, err := suite.repo.FindByID(context.Background(), 999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

func (suite *RepositoryTestSuite) TestUpdate() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	suite.mockRepo.On("Update", mock.Anything, int64(1), updates).Return(nil)

	err := suite.repo.Update(context.Background(), 1, updates)
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestDelete() {
	suite.mockRepo.On("Delete", mock.Anything, int64(1)).Return(nil)

	err := suite.repo.Delete(context.Background(), 1)
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestMarkForDeletion() {
	deleteAt := time.Now().AddDate(0, 0, 14)

	suite.mockRepo.On("MarkForDeletion", mock.Anything, int64(1), mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.repo.MarkForDeletion(context.Background(), 1, deleteAt)
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestDeleteExpiredParts() {
	now := time.Now()

	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.repo.DeleteExpiredParts(context.Background(), now)
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestFindAll() {
	expectedParts := []Part{
		{PartCore: PartCore{Name: "Part 1", Quantity: 1}},
		{PartCore: PartCore{Name: "Part 2", Quantity: 2}},
		{PartCore: PartCore{Name: "Part 3", Quantity: 3}},
	}

	suite.mockRepo.On("FindAll", mock.Anything).Return(expectedParts, nil)

	found, err := suite.repo.FindAll(context.Background())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 3)
}

func (suite *RepositoryTestSuite) TestBulkDelete() {
	ids := []int64{1, 2, 3}
	suite.mockRepo.On("BulkDelete", mock.Anything, ids).Return(nil)

	err := suite.repo.BulkDelete(context.Background(), ids)
	assert.NoError(suite.T(), err)
}

func (suite *RepositoryTestSuite) TestGetStatistics() {
	expected := StatisticsResponse{
		TotalParts:    100,
		TotalQuantity: 500,
		TotalValue:    15000.0,
	}
	suite.mockRepo.On("GetStatistics", mock.Anything).Return(expected, nil)

	stats, err := suite.repo.GetStatistics(context.Background())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 100, stats.TotalParts)
}

func (suite *RepositoryTestSuite) TestGetTotalEarnings() {
	suite.mockRepo.On("GetTotalEarnings", mock.Anything).Return(5000.0, nil)

	earnings, err := suite.repo.GetTotalEarnings(context.Background())
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5000.0, earnings)
}

func (suite *RepositoryTestSuite) TestGetSupplierCodes() {
	expected := []string{"SUP001", "SUP002"}
	suite.mockRepo.On("GetSupplierCodes", mock.Anything).Return(expected, nil)

	codes, err := suite.repo.GetSupplierCodes(context.Background())
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), codes, 2)
}

func (suite *RepositoryTestSuite) TestPartColumnsConsistency() {
	cols := strings.Split(partColumns, ",")
	assert.Equal(suite.T(), 47, len(cols), "partColumns must have exactly 47 columns")

	var foundPeriod, foundDate, foundVin bool
	for _, col := range cols {
		trimmed := strings.TrimSpace(col)
		if trimmed == "car_release_period" {
			foundPeriod = true
		}
		if trimmed == "car_release_date" {
			foundDate = true
		}
		if trimmed == "vin" {
			foundVin = true
		}
	}
	assert.True(suite.T(), foundPeriod, "partColumns must contain car_release_period")
	assert.True(suite.T(), foundDate, "partColumns must contain car_release_date")
	assert.True(suite.T(), foundVin, "partColumns must contain vin")
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}

func (m *MockPartRepository) RenameSeller(context.Context, int64, string) ([]Part, error) {
	return nil, nil
}
