package repository

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"auth-service/internal/domain"
)


type RepositoryTestSuite struct {
	suite.Suite
	mockUserRepo     *MockUserRepository
	mockActivityRepo *MockActivityLogRepository
}

func (suite *RepositoryTestSuite) SetupTest() {
	suite.mockUserRepo = new(MockUserRepository)
	suite.mockActivityRepo = new(MockActivityLogRepository)
}

func (suite *RepositoryTestSuite) TestUserCreate() {
	user := &domain.User{Email: "test@example.com", Name: "Test User", Role: "operator"}
	expected := &domain.User{ID: 1, Email: "test@example.com", Name: "Test User", Role: "operator"}

	suite.mockUserRepo.On("Create", user).Return(expected, nil)
	result, err := suite.mockUserRepo.Create(user)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.ID)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserFindByID() {
	expected := &domain.User{ID: 1, Email: "test@example.com", Name: "Test"}

	suite.mockUserRepo.On("FindByID", int64(1)).Return(expected, nil)
	result, err := suite.mockUserRepo.FindByID(1)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test@example.com", result.Email)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserFindByID_NotFound() {
	suite.mockUserRepo.On("FindByID", int64(999)).Return(nil, assert.AnError)
	result, err := suite.mockUserRepo.FindByID(999)

	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserFindByEmailOrName() {
	expected := &domain.User{ID: 1, Email: "user@example.com", Name: "username"}

	suite.mockUserRepo.On("FindByEmailOrName", "user@example.com").Return(expected, nil)
	result, err := suite.mockUserRepo.FindByEmailOrName("user@example.com")

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), result.ID)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserUpdate() {
	params := domain.UpdateUserParams{Name: "Updated Name", Email: "updated@example.com"}
	expected := &domain.User{ID: 1, Name: "Updated Name", Email: "updated@example.com"}

	suite.mockUserRepo.On("Update", int64(1), params).Return(expected, nil)
	result, err := suite.mockUserRepo.Update(1, params)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated Name", result.Name)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserDelete() {
	suite.mockUserRepo.On("Delete", int64(1)).Return(nil)
	err := suite.mockUserRepo.Delete(1)

	assert.NoError(suite.T(), err)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserFindAll() {
	expected := []domain.User{
		{ID: 1, Email: "user1@example.com"},
		{ID: 2, Email: "user2@example.com"},
	}

	suite.mockUserRepo.On("FindAll").Return(expected, nil)
	results, err := suite.mockUserRepo.FindAll()

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserExistsByEmailOrName_Exists() {
	suite.mockUserRepo.On("ExistsByEmailOrName", "test@example.com", "").Return(true, nil)
	exists, err := suite.mockUserRepo.ExistsByEmailOrName("test@example.com", "")

	assert.NoError(suite.T(), err)
	assert.True(suite.T(), exists)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestUserCountAll() {
	suite.mockUserRepo.On("CountAll").Return(5, nil)
	count, err := suite.mockUserRepo.CountAll()

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 5, count)
	suite.mockUserRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestActivityLogCreate() {
	log := &domain.UserActivityLog{
		UserID:       1,
		UserName:     "Test",
		Action:       "login",
		ResourceType: "auth",
	}
	expected := &domain.UserActivityLog{
		ID:           10,
		UserID:       1,
		UserName:     "Test",
		Action:       "login",
		ResourceType: "auth",
		CreatedAt:    time.Now(),
	}

	suite.mockActivityRepo.On("Create", log).Return(expected, nil)
	result, err := suite.mockActivityRepo.Create(log)

	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(10), result.ID)
	suite.mockActivityRepo.AssertExpectations(suite.T())
}

func (suite *RepositoryTestSuite) TestActivityLogFindWithFilters() {
	uid := int64(1)
	filters := domain.ActivityLogFilters{UserID: &uid, Limit: 10}
	expected := []domain.UserActivityLog{
		{ID: 1, UserID: 1, Action: "login"},
		{ID: 2, UserID: 1, Action: "logout"},
	}

	suite.mockActivityRepo.On("FindWithFilters", filters).Return(expected, nil)
	results, err := suite.mockActivityRepo.FindWithFilters(filters)

	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), results, 2)
	suite.mockActivityRepo.AssertExpectations(suite.T())
}

func TestRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(RepositoryTestSuite))
}
