package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// RepositoryTestSuite - набор тестов для репозитория
type RepositoryTestSuite struct {
	suite.Suite
	db   *gorm.DB
	repo PartRepository
}

// SetupTest - настройка перед каждым тестом
func (suite *RepositoryTestSuite) SetupTest() {
	// Создаем in-memory базу данных для тестов
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Миграция схемы
	err = db.AutoMigrate(&Part{})
	suite.Require().NoError(err)

	suite.db = db
	suite.repo = NewPartRepository(db)
}

// TearDownTest - очистка после каждого теста
func (suite *RepositoryTestSuite) TearDownTest() {
	sqlDB, _ := suite.db.DB()
	err := sqlDB.Close()
	if err != nil {
		return
	}
}

// TestCreate - тест создания запчасти
func (suite *RepositoryTestSuite) TestCreate() {
	part := &Part{
		Name:        "Test Part",
		Quantity:    10,
		Description: "Test description",
		Category:    "Test Category",
		Price:       100.0,
	}

	err := suite.repo.Create(part)
	assert.NoError(suite.T(), err)
	assert.NotZero(suite.T(), part.ID)
}

// TestFindByID - тест поиска по ID
func (suite *RepositoryTestSuite) TestFindByID() {
	// Создаем тестовую запчасть
	part := &Part{
		Name:     "Test Part",
		Quantity: 5,
		Price:    50.0,
	}
	err := suite.repo.Create(part)
	suite.Require().NoError(err)

	// Ищем созданную запчасть
	found, err := suite.repo.FindByID(part.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), found)
	assert.Equal(suite.T(), part.Name, found.Name)
	assert.Equal(suite.T(), part.Quantity, found.Quantity)
}

// TestFindByID_NotFound - тест поиска несуществующей запчасти
func (suite *RepositoryTestSuite) TestFindByID_NotFound() {
	found, err := suite.repo.FindByID(999)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestUpdate - тест обновления запчасти
func (suite *RepositoryTestSuite) TestUpdate() {
	// Создаем тестовую запчасть
	part := &Part{
		Name:     "Original Name",
		Quantity: 10,
		Price:    100.0,
	}
	err := suite.repo.Create(part)
	suite.Require().NoError(err)

	// Обновляем
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}
	err = suite.repo.Update(part.ID, updates)
	assert.NoError(suite.T(), err)

	// Проверяем обновление
	updated, err := suite.repo.FindByID(part.ID)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", updated.Name)
	assert.Equal(suite.T(), 20, updated.Quantity)
}

// TestDelete - тест удаления запчасти
func (suite *RepositoryTestSuite) TestDelete() {
	// Создаем тестовую запчасть
	part := &Part{
		Name:     "To Delete",
		Quantity: 1,
	}
	err := suite.repo.Create(part)
	suite.Require().NoError(err)

	// Удаляем
	err = suite.repo.Delete(part.ID)
	assert.NoError(suite.T(), err)

	// Проверяем, что не найдена
	found, err := suite.repo.FindByID(part.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestMarkForDeletion - тест отметки для удаления
func (suite *RepositoryTestSuite) TestMarkForDeletion() {
	// Создаем тестовую запчасть
	part := &Part{
		Name:     "To Mark",
		Quantity: 1,
	}
	err := suite.repo.Create(part)
	suite.Require().NoError(err)

	// Отмечаем для удаления
	deleteAt := time.Now().AddDate(0, 0, 14)
	err = suite.repo.MarkForDeletion(part.ID, deleteAt)
	assert.NoError(suite.T(), err)

	// Проверяем
	updated, err := suite.repo.FindByID(part.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), updated.ToDeleteAt)
}

// TestDeleteExpiredParts - тест удаления просроченных запчастей
func (suite *RepositoryTestSuite) TestDeleteExpiredParts() {
	// Создаем запчасть с прошедшей датой удаления
	pastTime := time.Now().AddDate(0, 0, -1)
	part := &Part{
		Name:       "Expired",
		Quantity:   1,
		ToDeleteAt: &pastTime,
	}
	err := suite.repo.Create(part)
	suite.Require().NoError(err)

	// Удаляем просроченные
	err = suite.repo.DeleteExpiredParts(time.Now())
	assert.NoError(suite.T(), err)

	// Проверяем, что удалена
	found, err := suite.repo.FindByID(part.ID)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), found)
}

// TestFindAll - тест получения всех запчастей
func (suite *RepositoryTestSuite) TestFindAll() {
	// Создаем несколько запчастей
	parts := []*Part{
		{Name: "Part 1", Quantity: 1},
		{Name: "Part 2", Quantity: 2},
		{Name: "Part 3", Quantity: 3},
	}

	for _, p := range parts {
		err := suite.repo.Create(p)
		suite.Require().NoError(err)
	}

	// Получаем все
	found, err := suite.repo.FindAll(suite.db.Model(&Part{}))
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), found, 3)
}

// TestRunSuite - запуск всех тестов
func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
